package service

import (
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
)

type Pseudonym struct {
	pseudonymStore *store.PseudonymCSV
}

func NewPseudonymService(pseudonymStore *store.PseudonymCSV) *Pseudonym {
	return &Pseudonym{
		pseudonymStore: pseudonymStore,
	}
}

func (p *Pseudonym) ListPseudonyms() ([]*model.Pseudonym, error) {
	pseudonyms, err := p.pseudonymStore.ListPseudonyms()
	if err != nil {
		return nil, fmt.Errorf("failed to list pseudonyms from store: %w", err)
	}
	return pseudonyms, nil
}
