package service

import (
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

type Facsimile struct {
	facsimileStore *store.FacsimileSQL
}

func NewFacsimileService(facsimileStore *store.FacsimileSQL) *Facsimile {
	return &Facsimile{
		facsimileStore: facsimileStore,
	}
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
