package ghwrapper

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (d *Downloader) fetchRawContent(url string) (content []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	h := d.httpHeaders()
	// Ask for raw content directly
	h.Set("Accept", "application/vnd.github.v3.raw")
	req.Header = h

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		return nil, &httpError{"GitHub rate limit hit. Set GITHUB_TOKEN", resp.StatusCode}
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, &httpError{fmt.Sprintf("GitHub API error for %s: %s (%s)", url, resp.Status, strings.TrimSpace(string(body))), resp.StatusCode}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Downloader) httpHeaders() http.Header {
	h := http.Header{}
	// Use standard media type for contents API and downloads
	h.Set("Accept", "application/vnd.github+json")
	h.Set("User-Agent", "gh-folder-downloader-go")

	if d.GitToken != "" {
		h.Set("Authorization", fmt.Sprintf("Bearer %s", d.GitToken))
	}
	return h
}
