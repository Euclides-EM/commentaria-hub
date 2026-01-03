package service

import (
	"fmt"
	"slices"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/tiendc/go-deepcopy"
)

// todo: add interfaces to all services

type Edition struct {
	m map[string]*model.Edition
}

func NewEditionService() *Edition {
	// todo: load from DB, not hardcoded + make sure to not create sync issues with the metadata csvs
	return &Edition{
		m: map[string]*model.Edition{
			"Paris_1615": {
				Meta: model.Meta{ID: "Paris_1615"},
				Facsimiles: []*model.Facsimile{
					{
						Meta:    model.Meta{ID: "1"},
						ScanURL: "https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1615.pdf",
					},
					{
						Meta:    model.Meta{ID: "2"},
						ScanURL: "https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1615.pdf",
					},
				},
			},
			"Paris_1598a": {
				Meta: model.Meta{ID: "Paris_1598a"},
				Facsimiles: []*model.Facsimile{
					{
						Meta:    model.Meta{ID: "1"},
						ScanURL: "https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1598a.pdf",
					},
					{
						Meta:    model.Meta{ID: "2"},
						ScanURL: "https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1598a.pdf",
					},
				},
			},
			"London_1570": {
				Meta: model.Meta{ID: "London_1570"},
				Facsimiles: []*model.Facsimile{
					{
						Meta:    model.Meta{ID: "1"},
						ScanURL: "https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/London_1570.pdf",
					},
				},
			},
		},
	}
}

// ListEditions returns a list of editions.
// For now, it returns a hardcoded edition with an optional facsimile.
func (e *Edition) ListEditions(expand []model.EditionExpandOptions, orderBy []model.EditionOrderByOptions) ([]*model.Edition, error) {
	eds := make([]*model.Edition, 0)
	for _, edition := range e.m {
		ed := &model.Edition{
			Meta: model.Meta{ID: edition.ID},
		}
		if slices.Contains(expand, model.EditionExpandFacsimiles) {
			facs := make([]*model.Facsimile, len(edition.Facsimiles))
			for i, fac := range edition.Facsimiles {
				var dst *model.Facsimile
				if err := deepcopy.Copy(&dst, &fac); err != nil {
					return nil, fmt.Errorf("failed to copy annotation: %w", err)
				}
				facs[i] = dst
			}
			ed.Facsimiles = facs
		}
		eds = append(eds, ed)
	}
	return eds, nil
}

func (e *Edition) GetFacsimile(editionKey, facsimileID string) (*model.Edition, *model.Facsimile, error) {
	allEditions, err := e.ListEditions([]model.EditionExpandOptions{model.EditionExpandFacsimiles}, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list editions: %w", err)
	}

	var targetEdition *model.Edition
	var targetFacsimile *model.Facsimile
	for _, ed := range allEditions {
		if ed.ID != editionKey {
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

func (e *Edition) UpdateFacsimile(key string, id string, f *model.Facsimile) (*model.Facsimile, error) {
	edition, ok := e.m[key]
	if !ok {
		return nil, fmt.Errorf("edition not found")
	}
	for i, fac := range edition.Facsimiles {
		if fac.ID == id {
			var dst *model.Facsimile
			if err := deepcopy.Copy(&dst, &f); err != nil {
				return nil, fmt.Errorf("failed to copy annotation: %w", err)
			}
			fac = dst
			e.m[key].Facsimiles[i] = fac
			return fac, nil
		}
	}
	return nil, fmt.Errorf("facsimile not found")
}
