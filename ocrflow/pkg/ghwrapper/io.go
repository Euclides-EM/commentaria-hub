package ghwrapper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
)

func (d *Downloader) fetchRawContent(ctx context.Context, url string) (content []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	h := d.httpHeaders(ctx)
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

func (d *Downloader) httpHeaders(ctx context.Context) http.Header {
	h := http.Header{}
	// Use standard media type for contents API and downloads
	h.Set("Accept", "application/vnd.github+json")
	h.Set("User-Agent", "gh-folder-downloader-go")

	// Check if token is in context first
	if token, ok := ctx.Value(httpwrapper.GitHubTokenKey).(string); ok && token != "" {
		h.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	} else if d.GitToken != "" {
		h.Set("Authorization", fmt.Sprintf("Bearer %s", d.GitToken))
	}
	return h
}
