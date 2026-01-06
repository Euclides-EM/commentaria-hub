package service

import (
	"fmt"
	"slices"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
)

// todo: add interfaces to all services

type Edition struct {
	editionStore   *store.EditionSQL
	facsimileStore *store.FacsimileSQL
}

func NewEditionService(editionStore *store.EditionSQL, facsimileStore *store.FacsimileSQL) *Edition {
	return &Edition{
		editionStore:   editionStore,
		facsimileStore: facsimileStore,
	}
}

// ListEditions returns a list of editions.
// For now, it returns a hardcoded edition with an optional facsimile.
func (e *Edition) ListEditions(expand []model.EditionExpandOptions, orderBy []model.EditionOrderByOptions) ([]*model.Edition, error) {
	eds, err := e.editionStore.ListEditions()
	if err != nil {
		return nil, fmt.Errorf("failed to list editions from store: %w", err)
	}
	for _, edition := range eds {
		if slices.Contains(expand, model.EditionExpandFacsimiles) {

			facs, err := e.facsimileStore.ListFacsimilesByEditionID(edition.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to list facsimiles from store: %w", err)
			}
			edition.Facsimiles = facs
		}
	}
	return eds, nil
}

func (e *Edition) GetFacsimile(editionKey, facsimileID string) (*model.Edition, *model.Facsimile, error) {
	ed, err := e.editionStore.GetEditionByID(editionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get edition from store: %w", err)
	}
	if ed == nil {
		return nil, nil, fmt.Errorf("edition with key %s not found", editionKey)
	}
	fac, err := e.facsimileStore.GetFacsimileByID(editionKey, facsimileID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get facsimile from store: %w", err)
	}
	if fac == nil {
		return nil, nil, fmt.Errorf("facsimile with id %s not found", facsimileID)
	}
	return ed, fac, nil
}
