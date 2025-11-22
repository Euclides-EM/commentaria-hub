package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
	"log"
	"slices"
)

// todo: add interfaces to all services

type Edition struct {
	m                map[string]*models.Edition
	GithubDownloader *ghwrapper.Downloader
	FacsimilesDir    string
}

func NewEditionService(facsimilesDir string, githubDownloader *ghwrapper.Downloader) *Edition {
	// todo: load from DB, not hardcoded + make sure to not create sync issues with the metadata csvs
	return &Edition{
		m: map[string]*models.Edition{
			"Paris_1615": {
				Key: "Paris_1615",
				Facsimiles: []*models.Facsimile{
					{
						ID:      "1",
						ScanURL: "https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1516.pdf",
						//ScanURL: "https://github.com/OCR-D/gt_structure_text/tree/main/data/alberti_pictura_1540",
					},
				},
			},
		},
		GithubDownloader: githubDownloader,
		FacsimilesDir:    facsimilesDir,
	}
}

// ListEditions returns a list of editions.
// For now, it returns a hardcoded edition with an optional facsimile.
func (e *Edition) ListEditions(expand []models.EditionExpandOptions, orderBy []models.EditionOrderByOptions) ([]*models.Edition, error) {
	eds := make([]*models.Edition, 0)
	for _, edition := range e.m {
		ed := &models.Edition{
			Key: edition.Key,
		}
		if slices.Contains(expand, models.EditionExpandFacsimiles) {
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
	allEditions, err := e.ListEditions([]models.EditionExpandOptions{models.EditionExpandFacsimiles}, nil)
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

	if !ghwrapper.IsGitHubTreeURL(targetFacsimile.ScanURL) {
		return fmt.Errorf("only GitHub tree URLs are supported currently")
	}

	localPath := fmt.Sprintf("%s/%s/%s.pdf", e.FacsimilesDir, editionKey, facsimileID)
	if err := e.GithubDownloader.DownloadRecursive(targetFacsimile.ScanURL, localPath); err != nil {
		return fmt.Errorf("failed to download facsimile: %w", err)
	}

	// todo: update DB with local path, currently it just happens implicitly in memory cause I use pointers
	e.UpdateFacsimile(editionKey, facsimileID, localPath)
	return nil
}

func (e *Edition) UpdateFacsimile(key string, id string, path string) {
	edition, ok := e.m[key]
	if !ok {
		return
	}
	for _, fac := range edition.Facsimiles {
		if fac.ID == id {
			fac.LocalPath = path
			return
		}
	}
}
