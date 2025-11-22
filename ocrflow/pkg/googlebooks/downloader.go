package googlebooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Download downloads the PDF from the given Google Books scan URL
// and saves it to destPath.
// For example, the scan URL can be: https://www.google.com/books/edition/Euclidis_elementorum_libri_XV/jEoRPuxe_pcC
// Then, the metadata about the book is fetched from Google Books API, from https://www.googleapis.com/books/v1/volumes/XIhmAAAAcAAJ
// The full documentation can be found here: https://developers.google.com/books/docs/v1/using#request_1
// Important note: Currently, this func is not working - Google books opens a CAPTCHA page instead of the actual PDF download.
// However, I am leaving this code here for future reference, in case we find a way to bypass the CAPTCHA or if we want to retrieve other metadata.
func Download(scanURL, destPath string) error {
	googleBookKey := extractGoogleBookKey(scanURL)
	if googleBookKey == "" {
		return fmt.Errorf("unsupported scan URL format")
	}

	httpClient := http.Client{}
	metaRes, err := httpClient.Get(
		fmt.Sprintf("https://www.googleapis.com/books/v1/volumes/%s", googleBookKey),
	)
	if err != nil {
		return fmt.Errorf("failed to fetch book metadata from Google Books API: %w", err)
	}
	defer metaRes.Body.Close()

	if metaRes.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response from Google Books API: %s", metaRes.Status)
	}

	var volume struct {
		AccessInfo struct {
			PDF struct {
				IsAvailable  bool   `json:"isAvailable"`
				DownloadLink string `json:"downloadLink"`
			} `json:"pdf"`
		} `json:"accessInfo"`
	}
	if err := json.NewDecoder(metaRes.Body).Decode(&volume); err != nil {
		return fmt.Errorf("failed to decode Google Books response: %w", err)
	}

	if !volume.AccessInfo.PDF.IsAvailable {
		return fmt.Errorf("pdf is not available for this volume")
	}

	if volume.AccessInfo.PDF.DownloadLink == "" {
		return fmt.Errorf("pdf marked as available but download link is empty")
	}

	pdfRes, err := httpClient.Get(volume.AccessInfo.PDF.DownloadLink)
	if err != nil {
		return fmt.Errorf("failed to download pdf: %w", err)
	}
	defer pdfRes.Body.Close()

	if pdfRes.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download pdf, status: %s", pdfRes.Status)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("failed to create facsimiles dir: %w", err)
	}

	out, err := os.Create(filepath.Dir(destPath))
	if err != nil {
		return fmt.Errorf("failed to create pdf file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, pdfRes.Body); err != nil {
		return fmt.Errorf("failed to write pdf to disk: %w", err)
	}
	return nil
}

func extractGoogleBookKey(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	if u.Hostname() != "www.google.com" {
		return ""
	}
	if !strings.Contains(u.Path, "/books/") {
		return ""
	}
	parts := strings.Split(u.Path, "/")
	return parts[len(parts)-1]
}
