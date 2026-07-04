package service

import (
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
)

type DiagramCrops struct {
	diagramsCropsStore *store.DiagramCropsStore
}

func (c *DiagramCrops) GetEditionDiagrams(key string) (*model.DiagramCrops, error) {
	return c.diagramsCropsStore.GetEditionDiagrams(key)
}

func NewDiagramCropsService(diagramsCropsStore *store.DiagramCropsStore) *DiagramCrops {
	return &DiagramCrops{
		diagramsCropsStore: diagramsCropsStore,
	}
}
