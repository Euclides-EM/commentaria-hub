package service

import (
	"fmt"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
)

type Shelfmark struct {
	editionStore *store.EditionCSV
}

func NewShelfmarkService(editionStore *store.EditionCSV) *Shelfmark {
	return &Shelfmark{editionStore: editionStore}
}

func (s *Shelfmark) ListShelfmarks(editionID string) ([]*model.EditionShelfmark, error) {
	if editionID == "" {
		return nil, fmt.Errorf("edition ID is required")
	}
	return s.editionStore.ListShelfmarks([]string{editionID})
}

func (s *Shelfmark) ListShelfmarksByEditionIDs(editionIDs []string) ([]*model.EditionShelfmark, error) {
	return s.editionStore.ListShelfmarks(editionIDs)
}

func (s *Shelfmark) ListAllShelfmarks() ([]*model.EditionShelfmark, error) {
	return s.editionStore.ListShelfmarks(nil)
}

func (s *Shelfmark) GetShelfmark(editionID, shelfmarkID string) (*model.EditionShelfmark, error) {
	if editionID == "" || shelfmarkID == "" {
		return nil, fmt.Errorf("edition ID and shelfmark ID are required")
	}
	sh, err := s.editionStore.GetShelfmark(editionID, shelfmarkID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, fmt.Errorf("shelfmark %s not found in edition %s", shelfmarkID, editionID)
	}
	return sh, nil
}

func (s *Shelfmark) UpsertShelfmark(editionID string, sh *model.EditionShelfmark) (*model.EditionShelfmark, error) {
	if editionID == "" {
		return nil, fmt.Errorf("edition ID is required")
	}
	return s.editionStore.UpsertShelfmark(editionID, sh)
}

func (s *Shelfmark) DeleteShelfmark(editionID, shelfmarkID string) error {
	if editionID == "" || shelfmarkID == "" {
		return fmt.Errorf("edition ID and shelfmark ID are required")
	}
	return s.editionStore.DeleteShelfmark(editionID, shelfmarkID)
}
