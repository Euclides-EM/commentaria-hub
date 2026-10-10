package service

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
)

type DiagramCrops struct {
	diagramsCropsStore     *store.DiagramCropsStore
	facsimilesDiagramsPath string
}

func (c *DiagramCrops) GetEditionDiagrams(key string) (*model.DiagramCrops, error) {
	return c.diagramsCropsStore.GetEditionDiagrams(key)
}

func (c *DiagramCrops) GetFacsimileDiagrams(facsimile *model.Facsimile, editionFacsimiles []*model.Facsimile) (*model.DiagramCrops, error) {
	if facsimile == nil {
		return &model.DiagramCrops{
			ImageURLsByName: map[string]string{},
			HasDiagrams:     false,
		}, nil
	}
	_, diagrams, err := c.resolveFacsimileDiagrams(facsimile, editionFacsimiles)
	if err != nil {
		return nil, err
	}
	if diagrams != nil {
		return diagrams, nil
	}
	return &model.DiagramCrops{
		Key:             facsimile.ID,
		ImageURLsByName: map[string]string{},
		HasDiagrams:     false,
	}, nil
}

func (c *DiagramCrops) resolveFacsimileDiagrams(facsimile *model.Facsimile, editionFacsimiles []*model.Facsimile) (string, *model.DiagramCrops, error) {
	keys := facsimileSpecificDiagramLookupKeys(facsimile)
	if facsimileEditionDiagramFallbackAllowed(facsimile, editionFacsimiles) {
		keys = append(keys, facsimile.EditionID)
	}
	for _, key := range keys {
		diagrams, err := c.diagramsCropsStore.GetEditionDiagrams(key)
		if err != nil {
			return "", nil, err
		}
		if diagrams.HasDiagrams || len(diagrams.ImageURLsByName) > 0 || len(diagrams.Volumes) > 0 {
			return key, diagrams, nil
		}
	}
	return "", nil, nil
}

func (c *DiagramCrops) ZipEditionDiagrams(key string) (string, func(io.Writer) error, error) {
	return c.zipDiagrams(key)
}

func (c *DiagramCrops) ZipFacsimileDiagrams(facsimile *model.Facsimile, editionFacsimiles []*model.Facsimile) (string, func(io.Writer) error, error) {
	key, _, err := c.resolveFacsimileDiagrams(facsimile, editionFacsimiles)
	if err != nil {
		return "", nil, err
	}
	if key == "" {
		return "", nil, fmt.Errorf("no diagram crops found for facsimile %s", facsimile.ID)
	}
	return c.zipDiagrams(key)
}

func (c *DiagramCrops) zipDiagrams(key string) (string, func(io.Writer) error, error) {
	diagramsDir, err := futils.LocalDirFromPathOrURL(c.facsimilesDiagramsPath)
	if err != nil {
		return "", nil, err
	}
	if diagramsDir == "" {
		return "", nil, fmt.Errorf("the facsimiles diagrams path must be an absolute local path or file:// URL")
	}
	relPaths, err := c.diagramsCropsStore.GetEditionDiagramCropRelPaths(key)
	if err != nil {
		return "", nil, err
	}
	if len(relPaths) == 0 {
		return "", nil, fmt.Errorf("no diagram crops found for %s", key)
	}
	write := func(w io.Writer) error {
		zw := zip.NewWriter(w)
		for _, relPath := range relPaths {
			if err := addFileToZip(zw, filesys.DiagramCropPath(diagramsDir, relPath), relPath); err != nil {
				return err
			}
		}
		return zw.Close()
	}
	return key + "-diagrams.zip", write, nil
}

func facsimileSpecificDiagramLookupKeys(facsimile *model.Facsimile) []string {
	keys := []string{}
	if name := strings.TrimSpace(facsimile.Name); name != "" {
		keys = append(keys, name)
	}
	if base := facsimileScanBasename(facsimile.ScanURL); base != "" {
		keys = append(keys, strings.TrimSuffix(base, filepath.Ext(base)))
	}
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, key)
	}
	return unique
}

func facsimileEditionDiagramFallbackAllowed(facsimile *model.Facsimile, editionFacsimiles []*model.Facsimile) bool {
	if facsimile == nil || strings.TrimSpace(facsimile.EditionID) == "" {
		return false
	}
	if len(editionFacsimiles) <= 1 {
		return true
	}
	mapped := []*model.Facsimile{}
	for _, candidate := range editionFacsimiles {
		if candidate != nil && strings.TrimSpace(candidate.ShelfmarkID) != "" {
			mapped = append(mapped, candidate)
		}
	}
	return len(mapped) == 1 && mapped[0].ID == facsimile.ID
}

func NewDiagramCropsService(diagramsCropsStore *store.DiagramCropsStore, facsimilesDiagramsPath string) *DiagramCrops {
	return &DiagramCrops{
		diagramsCropsStore:     diagramsCropsStore,
		facsimilesDiagramsPath: strings.TrimSpace(facsimilesDiagramsPath),
	}
}
