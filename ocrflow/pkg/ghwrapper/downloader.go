package ghwrapper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/avast/retry-go"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const apiBase = "https://api.github.com"
const githubHostname = "github.com"
const lfsPointerPrefix = "version https://git-lfs.github.com/spec/v1"

type Downloader struct {
	GitToken   string
	httpClient *http.Client
}

func NewDownloader(token string, timeout time.Duration) *Downloader {
	return &Downloader{
		GitToken: token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// DownloadRecursive downloads all files under the GitHub tree URL into destRoot.
func (d *Downloader) DownloadRecursive(treeURL, destRoot string) error {
	pu, err := parseGitHubTreeURL(treeURL)
	if err != nil {
		return err
	}

	destRootAbs, err := filepath.Abs(destRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve dest: %w", err)
	}

	return d.recurseDownload(pu, destRootAbs, "")
}

type ghEntry struct {
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
	SHA         string `json:"sha"` // The Git object SHA, needed for blob fallback
}

type httpError struct {
	msg        string
	statusCode int
}

func (e httpError) Error() string {
	return e.msg
}

func (d *Downloader) listDir(owner, repo, path, ref string) ([]ghEntry, *ghEntry, error) {
	apiPath := strings.TrimLeft(path, "/")
	u := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s",
		apiBase, owner, repo, url.PathEscape(apiPath), url.QueryEscape(ref))

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header = d.httpHeaders()

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, fmt.Errorf("not found: %s", u)
	}
	if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		return nil, nil, errors.New("GitHub rate limit hit. Set GITHUB_TOKEN")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, nil, fmt.Errorf("GitHub API error: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	// GitHub returns either an array for directories or a single object for a file
	if len(data) > 0 && data[0] == '[' {
		var entries []ghEntry
		if err := json.Unmarshal(data, &entries); err != nil {
			return nil, nil, err
		}
		return entries, nil, nil
	}

	var entry ghEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, nil, err
	}
	return nil, &entry, nil
}

func writeToFile(destPath string, reader io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	return err
}

// downloadFile attempts to download the file, handling LFS pointers and CDN 404 fallbacks.
func (d *Downloader) downloadFile(owner, repo, finalURL, sha, destPath string) error {
	if err := retry.Do(func() error {
		// First, get the file content
		contentData, err := d.fetchRawContent(finalURL)

		// If 404, try Git blob API as fallback (handles CDN cases)
		var httperr *httpError
		if err != nil && errors.As(err, &httperr) && httperr != nil && httperr.statusCode == http.StatusNotFound {
			contentData, err = d.fetchRawContent(fmt.Sprintf("%s/repos/%s/%s/git/blobs/%s", apiBase, owner, repo, sha))
		}

		if err != nil {
			return err
		}

		// Check if it's an LFS pointer
		lfsPtr, err := parseLfsPointer(contentData)
		if err != nil {
			// Not an LFS pointer, proceed to write the content.
			return writeToFile(destPath, bytes.NewReader(contentData))
		}

		// It's an LFS pointer, handle LFS download
		lfsLink, err := d.getLFSDownloadLink(owner, repo, lfsPtr.Oid, lfsPtr.Size)
		if err != nil {
			return fmt.Errorf("failed to get LFS download link for %s: %w", destPath, err)
		}
		return d.downloadLFS(lfsLink, destPath)
	}, retry.RetryIf(func(err error) bool {
		var httperr *httpError
		return err != nil && errors.As(err, &httperr) && (httperr.statusCode == 429 || httperr.statusCode == 502 || httperr.statusCode == 503 || httperr.statusCode == 504)
	}), retry.OnRetry(func(n uint, err error) {
		log.Printf("Failed to download %s (attempt %d), retrying: %v", destPath, n+1, err)
	}), retry.Attempts(5)); err != nil {
		return fmt.Errorf("failed to download %s: %s", destPath, err)
	}
	return nil
}

func (d *Downloader) recurseDownload(gt *ghURLTree, destRoot, relPath string) error {
	apiPath := strings.Trim(strings.Trim(gt.path+"/"+relPath, "/"), "/")

	entries, singleFile, err := d.listDir(gt.owner, gt.repo, apiPath, gt.ref)
	if err != nil {
		return err
	}

	// single file case
	if singleFile != nil && singleFile.Type == "file" {
		rel := strings.TrimPrefix(singleFile.Path, gt.path)
		rel = strings.TrimPrefix(rel, "/")
		dest, err := futils.SafeJoin(destRoot, rel)
		if err != nil {
			return err
		}
		// PASS SHA to downloadFile
		return d.downloadFile(gt.owner, gt.repo, singleFile.DownloadURL, singleFile.SHA, dest)
	}

	for _, e := range entries {
		switch e.Type {
		case "file":
			rel := strings.TrimPrefix(e.Path, gt.path)
			rel = strings.TrimPrefix(rel, "/")
			dest, err := futils.SafeJoin(destRoot, rel)
			if err != nil {
				return err
			}
			// PASS SHA to downloadFile
			if err := d.downloadFile(gt.owner, gt.repo, e.DownloadURL, e.SHA, dest); err != nil {
				return err
			}
		case "dir":
			subRel := strings.TrimPrefix(e.Path, gt.path)
			subRel = strings.TrimPrefix(subRel, "/")
			if err := d.recurseDownload(gt, destRoot, subRel); err != nil {
				return err
			}
		default:
			log.Printf("unrecognized entry type: %+v", e)
		}
	}

	return nil
}
