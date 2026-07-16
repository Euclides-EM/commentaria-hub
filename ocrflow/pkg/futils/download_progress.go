package futils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultDownloadProgressInterval = 10 * time.Second

func DownloadFileWithProgress(req *http.Request, dst string) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	progress := newFacsimileDownloadProgress(req.URL.String(), resp.ContentLength)
	if _, err := io.Copy(out, io.TeeReader(resp.Body, progress)); err != nil {
		return err
	}
	if err := out.Sync(); err != nil {
		return err
	}
	progress.complete()
	return nil
}

type downloadProgress struct {
	rawURL      string
	totalBytes  int64
	downloaded  int64
	startedAt   time.Time
	lastLogTime time.Time
}

func newFacsimileDownloadProgress(rawURL string, totalBytes int64) *downloadProgress {
	now := time.Now()
	return &downloadProgress{
		rawURL:      rawURL,
		totalBytes:  totalBytes,
		startedAt:   now,
		lastLogTime: now,
	}
}

func (p *downloadProgress) Write(data []byte) (int, error) {
	p.downloaded += int64(len(data))
	now := time.Now()
	if now.Sub(p.lastLogTime) >= defaultDownloadProgressInterval {
		p.log(now, false)
		p.lastLogTime = now
	}
	return len(data), nil
}

func (p *downloadProgress) complete() {
	p.log(time.Now(), true)
}

func (p *downloadProgress) log(now time.Time, complete bool) {
	action := "Downloading"
	if complete {
		action = "Downloaded"
	}
	elapsed := now.Sub(p.startedAt).Round(time.Millisecond)
	if p.totalBytes > 0 {
		percent := float64(p.downloaded) * 100 / float64(p.totalBytes)
		log.Printf("%s facsimile from %s: %.1f MiB / %.1f MiB (%.1f%%) in %s", action, p.rawURL, bytesToMiB(p.downloaded), bytesToMiB(p.totalBytes), percent, elapsed)
		return
	}
	log.Printf("%s facsimile from %s: %.1f MiB in %s", action, p.rawURL, bytesToMiB(p.downloaded), elapsed)
}

func bytesToMiB(size int64) float64 {
	return float64(size) / (1024 * 1024)
}
