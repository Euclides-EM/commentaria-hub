package ghwrapper

import (
	"bytes"
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
	"regexp"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.github.com"
const githubHostname = "github.com"
const lfsPointerPrefix = "version https://git-lfs.github.com/spec/v1"

// --- Downloader Core Struct ---

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

// --- GitHub API Types ---

type ghEntry struct {
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
	SHA         string `json:"sha"` // The Git object SHA, needed for blob fallback
}

// --- Git LFS Types ---

type lfsPointer struct {
	Oid  string
	Size int64
}

type lfsBatchRequest struct {
	Operation string `json:"operation"`
	Objects   []struct {
		Oid  string `json:"oid"`
		Size int64  `json:"size"`
	} `json:"objects"`
}

type lfsBatchResponse struct {
	Objects []struct {
		Oid   string `json:"oid"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
		Actions *struct {
			Download *struct {
				Href   string            `json:"href"`
				Header map[string]string `json:"header"`
			} `json:"download"`
		} `json:"actions"`
	} `json:"objects"`
}

// --- LFS Helper Functions ---

var lfsRegex = regexp.MustCompile(`oid sha256:([a-f0-9]{64})\s+size\s+(\d+)`)

func parseLfsPointer(data []byte) (*lfsPointer, error) {
	if !bytes.HasPrefix(data, []byte(lfsPointerPrefix)) {
		return nil, errors.New("not an LFS pointer file")
	}

	matches := lfsRegex.FindSubmatch(data)
	if len(matches) < 3 {
		return nil, fmt.Errorf("failed to parse LFS pointer: content incomplete")
	}

	oid := string(matches[1])
	sizeStr := string(matches[2])
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LFS size: %w", err)
	}

	return &lfsPointer{Oid: oid, Size: size}, nil
}

// getLFSDownloadLink uses the LFS Batch API to exchange an OID for a direct download URL.
func (d *Downloader) getLFSDownloadLink(owner, repo, oid string, size int64) (string, error) {
	// LFS API endpoint follows the format: https://github.com/<owner>/<repo>.git/info/lfs/objects/batch
	lfsApiUrl := fmt.Sprintf("https://%s/%s/%s.git/info/lfs/objects/batch", githubHostname, owner, repo)

	reqBody := lfsBatchRequest{
		Operation: "download",
		Objects: []struct {
			Oid  string `json:"oid"`
			Size int64  `json:"size"`
		}{
			{Oid: oid, Size: size},
		},
	}

	payload, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, lfsApiUrl, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	// LFS API uses a different Accept header
	req.Header = d.githubHeaders()
	req.Header.Set("Accept", "application/vnd.git-lfs+json")
	req.Header.Set("Content-Type", "application/vnd.git-lfs+json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return "", fmt.Errorf("LFS Batch API error: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
	}

	// Read the limited body into a byte slice because json.Unmarshal requires []byte.
	limitedBody := io.LimitReader(resp.Body, 1<<20) // Limit to 1MB
	data, err := io.ReadAll(limitedBody)
	if err != nil {
		return "", fmt.Errorf("failed to read LFS batch response body: %w", err)
	}

	var batchResp lfsBatchResponse
	if err := json.Unmarshal(data, &batchResp); err != nil {
		return "", fmt.Errorf("failed to parse LFS batch response: %w", err)
	}

	if len(batchResp.Objects) == 0 || batchResp.Objects[0].Actions == nil || batchResp.Objects[0].Actions.Download == nil {
		return "", fmt.Errorf("LFS Batch API did not return a download link for OID %s", oid)
	}

	return batchResp.Objects[0].Actions.Download.Href, nil
}

// downloadBlob fetches the raw content of a file using its Git SHA via the Blob API.
func (d *Downloader) downloadBlob(owner, repo, sha string) ([]byte, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/git/blobs/%s", apiBase, owner, repo, sha)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	h := d.githubHeaders()
	// Ask for raw content directly
	h.Set("Accept", "application/vnd.github.v3.raw")
	req.Header = h

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("GitHub Blob API error for SHA %s: %s (%s)", sha, resp.Status, strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// --- Download Logic ---

func (d *Downloader) githubHeaders() http.Header {
	h := http.Header{}
	// Use standard media type for contents API and downloads
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

// downloadFile attempts to download the file, handling LFS pointers and CDN 404 fallbacks.
func (d *Downloader) downloadFile(owner, repo, initialDownloadURL, sha, destPath string) error {
	// finalURL is used for the actual download attempt (can be initial URL or LFS link)
	var finalURL = initialDownloadURL

	if err := retry.Do(func() error {
		var contentData []byte
		var err error

		// Phase 1: Get the content data (LFS pointer or small file content)
		// Only perform Phase 1 if we are still using the initial raw/CDN URL.
		if finalURL == initialDownloadURL {
			// Attempt 1: Try the direct download URL (most common case, and usually fastest)
			req, err := http.NewRequest(http.MethodGet, finalURL, nil)
			if err != nil {
				return err
			}
			req.Header = d.githubHeaders()

			resp, err := d.httpClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				// *** FIX HERE: CDN 404 encountered, falling back to Blob API ***
				resp.Body.Close()
				log.Printf("Direct download failed with 404 for %s. Falling back to Blob API (SHA: %s).", destPath, sha)

				// Attempt 2: Fallback to Blob API using SHA
				contentData, err = d.downloadBlob(owner, repo, sha)
				if err != nil {
					// If Blob API also fails, then the file truly seems missing or inaccessible.
					return fmt.Errorf("failed to download content for %s: CDN 404 and Blob API failed: %w", destPath, err)
				}
			} else if resp.StatusCode == http.StatusOK {
				// Success with CDN URL
				contentData, err = io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
			} else {
				// Handle non-OK status codes (e.g., 403, 5xx)
				if resp.StatusCode == 429 || resp.StatusCode == 502 ||
					resp.StatusCode == 503 || resp.StatusCode == 504 {
					return fmt.Errorf("transient error: http status %d", resp.StatusCode)
				}
				if resp.StatusCode == 403 && resp.Header.Get("X-RateLimit-Remaining") == "0" {
					return errors.New("GitHub rate limit hit. Set GITHUB_TOKEN")
				}

				body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
				return fmt.Errorf("failed to download %s (HTTP %d): %s", finalURL, resp.StatusCode, strings.TrimSpace(string(body)))
			}

			// Phase 2: LFS Check (only performed once using the initial content)
			if ptr, err := parseLfsPointer(contentData); err == nil {
				// LFS Pointer detected!
				log.Printf("LFS pointer detected for %s. Initiating LFS download protocol.", destPath)

				// Step 3: Get the actual download link from LFS API
				lfsLink, err := d.getLFSDownloadLink(owner, repo, ptr.Oid, ptr.Size)
				if err != nil {
					return fmt.Errorf("failed to get LFS download link for %s: %w", destPath, err)
				}

				// Update the URL and re-trigger the entire retry loop to download the actual file
				finalURL = lfsLink
				return errors.New("LFS_RETRY_TRIGGERED") // Custom error to force a retry with the new URL
			}

			// If it's not an LFS pointer, we have the final content. Write it and succeed.
			// (This handles small, non-LFS files downloaded successfully via CDN or Blob API)
			if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
				return err
			}
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()

			if _, err := io.Copy(out, bytes.NewReader(contentData)); err != nil {
				return err
			}
			return nil // Success with non-LFS file
		}

		// --- Execution continues here only if finalURL is the LFS link ---

		// Step 4: Download the large LFS file using the obtained LFS link
		req, err := http.NewRequest(http.MethodGet, finalURL, nil)
		if err != nil {
			return err
		}
		// NOTE: LFS download links may not need the GitToken, but it doesn't hurt to include it.
		req.Header = d.githubHeaders()

		resp, err := d.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for transient errors in the final LFS download
		if resp.StatusCode == 429 || resp.StatusCode == 502 ||
			resp.StatusCode == 503 || resp.StatusCode == 504 {
			return fmt.Errorf("transient error: http status %d", resp.StatusCode)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
			return fmt.Errorf("failed to download LFS content %s (HTTP %d): %s", finalURL, resp.StatusCode, strings.TrimSpace(string(body)))
		}

		// Write the final LFS content to file
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}
		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		return err // Return nil on success

	}, retry.RetryIf(func(err error) bool {
		return strings.Contains(err.Error(), "transient error:") || err.Error() == "LFS_RETRY_TRIGGERED"
	}), retry.OnRetry(func(n uint, err error) {
		// Log LFS trigger only once
		if err.Error() == "LFS_RETRY_TRIGGERED" {
			log.Printf("Successfully obtained LFS download link for %s. Re-running download (attempt %d).", destPath, n+1)
			return
		}
		log.Printf("Failed to download %s (attempt %d), retrying: %v", destPath, n+1, err)
	}), retry.Attempts(5)); err != nil {
		// If the LFS retry trigger is the final error, it means the second download failed.
		if err.Error() == "LFS_RETRY_TRIGGERED" {
			return fmt.Errorf("failed to complete LFS download after retrieving link: check network or permissions")
		}
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
		dest, err := safeJoin(destRoot, rel)
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
			dest, err := safeJoin(destRoot, rel)
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
