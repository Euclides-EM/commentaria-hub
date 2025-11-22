package ghwrapper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

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
	req.Header = d.httpHeaders()
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

func (d *Downloader) downloadLFS(finalURL, destPath string) error {
	req, err := http.NewRequest(http.MethodGet, finalURL, nil)
	if err != nil {
		return err
	}
	req.Header = d.httpHeaders()

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return httpError{fmt.Sprintf("failed to download LFS content %s (HTTP %d): %s", finalURL, resp.StatusCode, strings.TrimSpace(string(body))), resp.StatusCode}
	}
	return writeToFile(destPath, resp.Body)
}
