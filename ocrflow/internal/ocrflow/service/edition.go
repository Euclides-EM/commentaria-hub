package service

import "github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"

type Edition struct {
	m map[string]*models.Edition
}

func NewEditionService() *Edition {
	return &Edition{
		m: map[string]*models.Edition{
			"Paris_1615": {
				Key: "Paris_1615",
				Facsimiles: []*models.Facsimile{
					{
						ScanURL: "https://www.google.com/books/edition/Les_quinze_livres_des_Elements_d_Euclide/XIhmAAAAcAAJ",
					},
				},
			},
		},
	}
}

// ListEditions returns a list of editions.
// For now, it returns a hardcoded edition with an optional facsimile.
func (e *Edition) ListEditions(includeFacsimiles bool) ([]*models.Edition, error) {
	eds := make([]*models.Edition, 0)
	for _, edition := range e.m {
		ed := &models.Edition{
			Key: edition.Key,
		}
		if includeFacsimiles {
			facs := make([]*models.Facsimile, len(edition.Facsimiles))
			for i, fac := range edition.Facsimiles {
				facs[i] = &models.Facsimile{
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

// https://www.googleapis.com/books/v1/volumes/XIhmAAAAcAAJ
// https://www.google.com/books/edition/Euclidis_elementorum_libri_XV/jEoRPuxe_pcC
// https://developers.google.com/books/docs/v1/using#request_1
