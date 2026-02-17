package service

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
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
