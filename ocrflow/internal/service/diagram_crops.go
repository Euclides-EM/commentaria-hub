package service

import (
	"path/filepath"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
)

type DiagramCrops struct {
	diagramsCropsStore *store.DiagramCropsStore
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
	for _, key := range facsimileSpecificDiagramLookupKeys(facsimile) {
		diagrams, err := c.diagramsCropsStore.GetEditionDiagrams(key)
		if err != nil {
			return nil, err
		}
		if diagrams.HasDiagrams || len(diagrams.ImageURLsByName) > 0 || len(diagrams.Volumes) > 0 {
			return diagrams, nil
		}
	}
	if facsimileEditionDiagramFallbackAllowed(facsimile, editionFacsimiles) {
		diagrams, err := c.diagramsCropsStore.GetEditionDiagrams(facsimile.EditionID)
		if err != nil {
			return nil, err
		}
		if diagrams.HasDiagrams || len(diagrams.ImageURLsByName) > 0 || len(diagrams.Volumes) > 0 {
			return diagrams, nil
		}
	}
	return &model.DiagramCrops{
		Key:             facsimile.ID,
		ImageURLsByName: map[string]string{},
		HasDiagrams:     false,
	}, nil
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

func NewDiagramCropsService(diagramsCropsStore *store.DiagramCropsStore) *DiagramCrops {
	return &DiagramCrops{
		diagramsCropsStore: diagramsCropsStore,
	}
}
