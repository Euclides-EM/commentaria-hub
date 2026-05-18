package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/samber/lo"
)

type Facsimile struct {
	facsimileStore   *store.FacsimileSQL
	ghDownloader     *ghwrapper.Wrapper
	facsimileRepoURL string
	facsimilesPDFDir string
}

func NewFacsimileService(facsimileStore *store.FacsimileSQL, downloader *ghwrapper.Wrapper, facsimileRepoURL, facsimilesPDFDir string) *Facsimile {
	return &Facsimile{
		facsimileStore:   facsimileStore,
		ghDownloader:     downloader,
		facsimileRepoURL: facsimileRepoURL,
		facsimilesPDFDir: strings.TrimSpace(facsimilesPDFDir),
	}
}

func (e *Facsimile) ListFacsimiles(editionIDs []string) ([]*model.Facsimile, error) {
	return e.facsimileStore.ListFacsimiles(editionIDs)
}

func (e *Facsimile) GetFacsimile(facsimileID string) (*model.Facsimile, error) {
	fac, err := e.facsimileStore.GetFacsimileByID(facsimileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile from store: %w", err)
	}
	if fac == nil {
		return nil, fmt.Errorf("facsimile with id %s not found", facsimileID)
	}
	return fac, nil
}

func (e *Facsimile) CreateFacsimile(f *model.Facsimile) (*model.Facsimile, error) {
	f.ID = idgen.GenerateID(store.FacsimileIDPrefix)
	f.CreatedAt = time.Now()
	f.UpdatedAt = f.CreatedAt
	return e.facsimileStore.InsertFacsimile(f)
}

func (e *Facsimile) UpdateFacsimile(f *model.Facsimile) (*model.Facsimile, error) {
	existing, err := e.facsimileStore.GetFacsimileByID(f.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("facsimile with id %s not found", f.ID)
	}
	return e.facsimileStore.UpdateFacsimile(f)
}

func (e *Facsimile) UpdateFromGithubRepo() error {
	files, err := e.ghDownloader.ListFiles(context.Background(), e.facsimileRepoURL)
	if err != nil {
		return fmt.Errorf("failed to list facsimiles: %w", err)
	}
	var keys []string
	for _, file := range files {
		if filepath.Ext(file) == ".pdf" {
			keys = append(keys, strings.TrimSuffix(file, ".pdf"))
		}
	}
	existingFacsimiles, err := e.ListFacsimiles(nil)
	if err != nil {
		return fmt.Errorf("failed to list existing facsimiles: %w", err)
	}
	existingFacsimilesKeys := lo.Map(existingFacsimiles, func(f *model.Facsimile, _ int) string {
		return f.EditionID
	})
	facsimilesToAdd := lo.Filter(keys, func(key string, _ int) bool {
		return !slices.Contains(existingFacsimilesKeys, key)
	})
	log.Printf("adding %d new facsimiles from github repo to db", len(facsimilesToAdd))
	for _, key := range facsimilesToAdd {
		u, _ := url.Parse(e.facsimileRepoURL)
		u.Path = path.Join(u.Path, fmt.Sprintf("%s.pdf", key))
		newFacsimile := &model.Facsimile{
			EditionID: key,
			ScanURL:   u.String(),
		}
		if _, err := e.CreateFacsimile(newFacsimile); err != nil {
			log.Printf("failed to create facsimile for %s: %v", key, err)
		}
	}
	return nil
}

func (e *Facsimile) UpdateFromConfiguredSource() error {
	if e.facsimilesPDFDir != "" {
		return e.UpdateFromLocalDir(e.facsimilesPDFDir)
	}
	return e.UpdateFromGithubRepo()
}

func (e *Facsimile) UpdateFromLocalDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to list local facsimiles in %s: %w", dir, err)
	}

	existingFacsimiles, err := e.ListFacsimiles(nil)
	if err != nil {
		return fmt.Errorf("failed to list existing facsimiles: %w", err)
	}
	existingByEditionID := lo.SliceToMap(existingFacsimiles, func(f *model.Facsimile) (string, *model.Facsimile) {
		return f.EditionID, f
	})

	added := 0
	updated := 0
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".pdf" {
			continue
		}
		editionID := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		pdfURL, err := futils.LocalFilePathToURL(filepath.Join(dir, entry.Name()))
		if err != nil {
			log.Printf("failed to build local file URL for %s: %v", entry.Name(), err)
			continue
		}

		if existing := existingByEditionID[editionID]; existing != nil {
			if existing.ScanURL == pdfURL {
				continue
			}
			existing.ScanURL = pdfURL
			if _, err := e.UpdateFacsimile(existing); err != nil {
				log.Printf("failed to update local facsimile for %s: %v", editionID, err)
				continue
			}
			updated++
			continue
		}

		if _, err := e.CreateFacsimile(&model.Facsimile{
			EditionID: editionID,
			ScanURL:   pdfURL,
		}); err != nil {
			log.Printf("failed to create local facsimile for %s: %v", editionID, err)
			continue
		}
		added++
	}

	log.Printf("local facsimile sync from %s added %d and updated %d facsimiles", dir, added, updated)
	return nil
}
