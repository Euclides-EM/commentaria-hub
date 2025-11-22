package ghwrapper

import (
	"encoding/json"
	"errors"
	"fmt"
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

// DownloadGitHubTree downloads all files under the GitHub tree URL into destRoot.
// URL must look like:
//
//	https://github.com/<owner>/<repo>/tree/<ref>/<path...>
func (d *Downloader) DownloadGitHubTree(treeURL, destRoot string) error {
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
}

func (d *Downloader) githubHeaders() http.Header {
	h := http.Header{}
	h.Set("Accept", "application/vnd.github+json")
	h.Set("User-Agent", "gh-folder-downloader-go")

	if d.GitToken != "" {
		h.Set("Authorization", fmt.Sprintf("Bearer %s", d.GitToken))
	}
	return h
}

func (d *Downloader) listDir(owner, repo, path, ref string) ([]ghEntry, *ghEntry, error) {
	apiPath := strings.TrimLeft(path, "/")
	u := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s",
		apiBase, owner, repo, url.PathEscape(apiPath), url.QueryEscape(ref))

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header = d.githubHeaders()

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

func (d *Downloader) downloadFile(downloadURL, destPath string) error {
	if err := retry.Do(func() error {
		req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
		if err != nil {
			return err
		}
		req.Header = d.githubHeaders()

		resp, err := d.httpClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
				return err
			}
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			return err
		}

		// transient errors
		if resp.StatusCode == 429 || resp.StatusCode == 502 ||
			resp.StatusCode == 503 || resp.StatusCode == 504 {
			resp.Body.Close()
			return fmt.Errorf("transient error: http status %d", resp.StatusCode)
		}

		if resp.StatusCode == 403 && resp.Header.Get("X-RateLimit-Remaining") == "0" {
			resp.Body.Close()
			return errors.New("GitHub rate limit hit. Set GITHUB_TOKEN")
		}

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		resp.Body.Close()
		return fmt.Errorf("failed to download %s (HTTP %d): %s", downloadURL, resp.StatusCode, strings.TrimSpace(string(body)))
	}, retry.RetryIf(func(err error) bool {
		return strings.Contains(err.Error(), "transient error:")
	}), retry.OnRetry(func(n uint, err error) {
		log.Printf("failed to download %s (attempt %d), retrying: %v", downloadURL, n, err)
	})); err != nil {
		return fmt.Errorf("failed to download %s: %s", downloadURL, err)
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
		dest, err := safeJoin(destRoot, rel)
		if err != nil {
			return err
		}
		return d.downloadFile(singleFile.DownloadURL, dest)
	}

	for _, e := range entries {
		switch e.Type {
		case "file":
			rel := strings.TrimPrefix(e.Path, gt.path)
			rel = strings.TrimPrefix(rel, "/")
			dest, err := safeJoin(destRoot, rel)
			if err != nil {
				return err
			}
			if err := d.downloadFile(e.DownloadURL, dest); err != nil {
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

func safeJoin(destRoot, relPath string) (string, error) {
	rootAbs, err := filepath.Abs(destRoot)
	if err != nil {
		return "", err
	}
	dest := filepath.Join(rootAbs, filepath.FromSlash(relPath))
	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return "", err
	}
	// ensure dest is under root
	if destAbs != rootAbs && !strings.HasPrefix(destAbs, rootAbs+string(filepath.Separator)) {
		return "", fmt.Errorf("refusing to write outside destination: %s", destAbs)
	}
	return destAbs, nil
}
