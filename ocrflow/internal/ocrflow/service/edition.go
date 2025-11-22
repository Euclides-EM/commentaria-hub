package service

import (
	"encoding/json"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// todo: add interfaces to all services

type Edition struct {
	m             map[string]*models.Edition
	FacsimilesDir string
}

func NewEditionService(facsimilesDir string) *Edition {
	// todo: load from DB, not hardcoded + make sure to not create sync issues with the metadata csvs
	return &Edition{
		FacsimilesDir: facsimilesDir,
		m: map[string]*models.Edition{
			"Paris_1615": {
				Key: "Paris_1615",
				Facsimiles: []*models.Facsimile{
					{
						ID:      "1",
						ScanURL: "https://www.google.com/books/edition/Les_quinze_livres_des_Elements_d_Euclide/XIhmAAAAcAAJ",
					},
				},
			},
		},
	}
}

// ListEditions returns a list of editions.
// For now, it returns a hardcoded edition with an optional facsimile.
func (e *Edition) ListEditions(includeFacsimiles bool) ([]*models.Edition, error) {
	eds := make([]*models.Edition, 0)
	for _, edition := range e.m {
		ed := &models.Edition{
			Key: edition.Key,
		}
		if includeFacsimiles {
			facs := make([]*models.Facsimile, len(edition.Facsimiles))
			for i, fac := range edition.Facsimiles {
				facs[i] = &models.Facsimile{
					ID:        fac.ID,
					ScanURL:   fac.ScanURL,
					LocalPath: fac.LocalPath,
				}
			}
			ed.Facsimiles = facs
		}
		eds = append(eds, ed)
	}
	return eds, nil
}

func (e *Edition) DownloadFacsimile(editionKey, facsimileID string, forceRedownload bool) error {
	allEditions, err := e.ListEditions(true)
	if err != nil {
		return fmt.Errorf("failed to list editions: %w", err)
	}

	var targetEdition *models.Edition
	var targetFacsimile *models.Facsimile
	for _, ed := range allEditions {
		if ed.Key != editionKey {
			continue
		}
		for _, fac := range ed.Facsimiles {
			if fac.ID == facsimileID {
				targetEdition = ed
				targetFacsimile = fac
				break
			}
		}
	}

	if targetEdition == nil || targetFacsimile == nil {
		// todo: add error handler with 404 response (currently returns 500)
		return fmt.Errorf("edition or facsimile not found")
	}

	if !forceRedownload && targetFacsimile.LocalPath != "" {
		log.Printf("facsimile already downloaded at %s, skipping", targetFacsimile.LocalPath)
		return nil
	}

	if targetFacsimile.ScanURL == "" {
		return fmt.Errorf("facsimile has no scan URL")
	}

	googleBookKey := extractGoogleBookKey(targetFacsimile.ScanURL)
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

	if err := os.MkdirAll(path.Join(e.FacsimilesDir, "pdfs", editionKey), 0o755); err != nil {
		return fmt.Errorf("failed to create facsimiles dir: %w", err)
	}

	destPath := filepath.Join(e.FacsimilesDir, "pdfs", editionKey, fmt.Sprintf("%s.pdf", facsimileID))

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create pdf file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, pdfRes.Body); err != nil {
		return fmt.Errorf("failed to write pdf to disk: %w", err)
	}

	// todo: update DB with local path, currently it just happens implicitly in memory cause I use pointers
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

// https://www.googleapis.com/books/v1/volumes/XIhmAAAAcAAJ
// https://www.google.com/books/edition/Euclidis_elementorum_libri_XV/jEoRPuxe_pcC
// https://developers.google.com/books/docs/v1/using#request_1
