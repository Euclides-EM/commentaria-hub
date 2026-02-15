package service

import (
	"fmt"
	"mime"
	"mime/multipart"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

// todo: add interfaces to all services

type Edition struct {
	editionStore   *store.EditionCSV
	facsimileStore *store.FacsimileSQL
}

func NewEditionService(editionStore *store.EditionCSV, facsimileStore *store.FacsimileSQL) *Edition {
	return &Edition{
		editionStore:   editionStore,
		facsimileStore: facsimileStore,
	}
}

// ListEditions returns a paginated list of editions, optionally filtered by corpus.
func (e *Edition) ListEditions(filter func(e any) bool, orderBy func(e1, e2 any) int, offset, limit int) (any, error) {
	items, total, err := e.editionStore.ListEditions(filter, orderBy, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list editions from store: %w", err)
	}
	return &model.EditionListResult{Items: items, Total: total, Offset: offset, Limit: limit}, nil
}

func (e *Edition) CreateEdition(ed *model.Edition, login string) (*model.Edition, error) {
	if ed.Key == "" {
		ed.Key = idgen.GenerateID("ed")
	}
	existing, err := e.GetEditionByID(ed.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing edition: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("edition with key %s already exists", ed.Key)
	}
	if err := e.editionStore.UpsertEdition(ed, login); err != nil {
		return nil, fmt.Errorf("failed to create edition: %w", err)
	}
	return e.GetEditionByID(ed.Key)
}

func (e *Edition) UpdateEdition(m *model.Edition, login string) (*model.Edition, error) {
	// Check if edition exists
	existing, err := e.GetEditionByID(m.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing edition: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("edition with key %s does not exist", m.Key)
	}
	if err := e.editionStore.UpsertEdition(m, login); err != nil {
		return nil, fmt.Errorf("failed to update edition: %w", err)
	}
	return e.GetEditionByID(m.Key)
}

func (e *Edition) UpdateNotes(key, note string) (*model.Edition, error) {
	if err := e.editionStore.UpdateNotes(key, note); err != nil {
		return nil, fmt.Errorf("failed to update notes for edition %s: %w", key, err)
	}
	return e.GetEditionByID(key)
}

func (e *Edition) GetEditionByID(key string) (*model.Edition, error) {
	ed, err := e.editionStore.GetEditionByID(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get edition by ID: %w", err)
	}
	return ed, nil
}

func (e *Edition) UploadImage(key string, typ string, file multipart.File, header *multipart.FileHeader) (*model.ImageUpload, error) {
	ext := strings.TrimPrefix(strings.ToLower(mime.TypeByExtension(header.Filename)), "image/")
	if ext == "" {
		return nil, fmt.Errorf("unable to determine file extension for uploaded image")
	}
	return e.editionStore.UploadImage(key, typ, ext, file)
}

func (e *Edition) DeleteEdition(key string) error {
	// Check if edition exists
	existing, err := e.GetEditionByID(key)
	if err != nil {
		return fmt.Errorf("failed to check for existing edition: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("edition with key %s does not exist", key)
	}
	// Delete associated facsimiles
	facs, err := e.facsimileStore.ListFacsimilesByEditionID(key)
	if err != nil {
		return fmt.Errorf("failed to list facsimiles for edition %s: %w", key, err)
	}
	for _, fac := range facs {
		if err := e.facsimileStore.DeleteFacsimile(fac.ID); err != nil {
			return fmt.Errorf("failed to delete facsimile %s for edition %s: %w", fac.ID, key, err)
		}
	}
	// Delete edition
	if err := e.editionStore.DeleteEdition(key); err != nil {
		return fmt.Errorf("failed to delete edition %s: %w", key, err)
	}
	return nil
}
