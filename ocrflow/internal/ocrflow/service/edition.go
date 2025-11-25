package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/krakenwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"log"
	"os"
	"path"
	"slices"
)

// todo: add interfaces to all services

type Edition struct {
	m                map[string]*models.Edition
	annotationsM     map[string]*models.Annotation
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
						ID:           "1",
						ScanURL:      "https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1615.pdf",
						PDFLocalPath: "./facsimiles/Paris_1615/1.pdf",
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
				facs[i] = fac.DeepCopy()
			}
			ed.Facsimiles = facs
		}
		eds = append(eds, ed)
	}
	return eds, nil
}

func (e *Edition) GetFacsimile(editionKey, facsimileID string) (*models.Edition, *models.Facsimile, error) {
	allEditions, err := e.ListEditions([]models.EditionExpandOptions{models.EditionExpandFacsimiles}, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list editions: %w", err)
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
		return nil, nil, fmt.Errorf("edition or facsimile not found")
	}
	return targetEdition, targetFacsimile, nil
}

func (e *Edition) DownloadFacsimile(editionKey, facsimileID string, forceRedownload bool) (*models.Facsimile, error) {
	_, targetFacsimile, err := e.GetFacsimile(editionKey, facsimileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}

	if !forceRedownload && targetFacsimile.PDFLocalPath != "" {
		log.Printf("facsimile already downloaded at %s, skipping", targetFacsimile.PDFLocalPath)
		return targetFacsimile, nil
	}

	if targetFacsimile.ScanURL == "" {
		return nil, fmt.Errorf("facsimile has no scan URL")
	}

	if !ghwrapper.IsGitHubTreeURL(targetFacsimile.ScanURL) {
		return nil, fmt.Errorf("only GitHub tree URLs are supported currently")
	}

	localPath := fmt.Sprintf("%s/%s/%s.pdf", e.FacsimilesDir, editionKey, facsimileID)
	if err := e.GithubDownloader.DownloadRecursive(targetFacsimile.ScanURL, localPath); err != nil {
		return nil, fmt.Errorf("failed to download facsimile: %w", err)
	}

	// todo: update DB with local path, currently it just happens implicitly in memory cause I use pointers
	return e.UpdateFacsimile(editionKey, facsimileID, &models.Facsimile{
		ID:           targetFacsimile.ID,
		ScanURL:      targetFacsimile.ScanURL,
		PDFLocalPath: localPath,
	})
}

func (e *Edition) PreProcessFacsimile(editionKey, facsimileID string, force bool) (*models.Facsimile, error) {
	_, targetFacsimile, err := e.GetFacsimile(editionKey, facsimileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}
	if !force && targetFacsimile.JPGsLocalDir != "" {
		log.Printf("facsimile already pre-processed at %s, skipping", targetFacsimile.JPGsLocalDir)
		return targetFacsimile, nil
	}
	if targetFacsimile.PDFLocalPath == "" {
		return nil, fmt.Errorf("facsimile has no local PDF path, download it first")
	}
	jpgsDir := fmt.Sprintf("%s/%s/%s_jpgs", e.FacsimilesDir, editionKey, facsimileID)
	if err := formatcov.PDF2JPGs(targetFacsimile.PDFLocalPath, jpgsDir); err != nil {
		return nil, fmt.Errorf("failed to pre-process facsimile: %w", err)
	}
	f := targetFacsimile.DeepCopy()
	f.JPGsLocalDir = jpgsDir
	return e.UpdateFacsimile(editionKey, facsimileID, f)
}

func (e *Edition) AnnotateFacsimile(a *models.Annotation) (*models.Annotation, error) {
	_, targetFacsimile, err := e.GetFacsimile(a.EditionKey, a.FacsimileId)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}
	if targetFacsimile.JPGsLocalDir == "" {
		return nil, fmt.Errorf("facsimile has not been pre-processed yet")
	}
	a.ID = idgen.GenerateID()
	outDir := fmt.Sprintf("%s/%s/%s_annotations/%s", e.FacsimilesDir, a.EditionKey, a.FacsimileId, a.ID)
	filenames := []string{}
	pages, err := pagesparser.Parse(a.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages: %w", err)
	}
	for _, page := range pages {
		filename := fmt.Sprintf("page-%04d.jpg", page)
		if _, err := os.Stat(path.Join(outDir, filename)); err != nil {
			return nil, fmt.Errorf("no such file %s in existing dataset", filename)
		}
		filenames = append(filenames, filename)
	}
	// inputDir string, outputDir string, krakenModel string, filenames []string
	if err := krakenwrapper.Recognize(targetFacsimile.JPGsLocalDir, outDir, a.Model, filenames); err != nil {
		return nil, fmt.Errorf("failed to annotate facsimile: %w", err)
	}
	a.AnnotatedLocalDir = outDir
	return e.UpdateAnnotation(a)
}

func (e *Edition) UpdateFacsimile(key string, id string, f *models.Facsimile) (*models.Facsimile, error) {
	edition, ok := e.m[key]
	if !ok {
		return nil, fmt.Errorf("edition not found")
	}
	for i, fac := range edition.Facsimiles {
		if fac.ID == id {
			fac = f.DeepCopy()
			e.m[key].Facsimiles[i] = fac
			return fac, nil
		}
	}
	return nil, fmt.Errorf("facsimile not found")
}

func (e *Edition) UpdateAnnotation(a *models.Annotation) (*models.Annotation, error) {
	e.annotationsM[a.ID] = a.DeepCopy()
	return a, nil
}
