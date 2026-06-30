package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/samber/lo"
)

// todo: add interfaces to all services

var ErrEditionNotFound = errors.New("edition not found")

const subjectCategoriesFeatureID = "m_classifier"

type Edition struct {
	editionStore       *store.EditionCSV
	facsimileStore     *store.FacsimileSQL
	featureResultStore *store.FeatureResultSql
}

func NewEditionService(editionStore *store.EditionCSV, facsimileStore *store.FacsimileSQL, featureResultStore *store.FeatureResultSql) *Edition {
	return &Edition{
		editionStore:       editionStore,
		facsimileStore:     facsimileStore,
		featureResultStore: featureResultStore,
	}
}

// ListEditions returns a paginated list of editions, optionally filtered by corpus.
func (e *Edition) ListEditions(filter func(e any) bool, orderBy func(e1, e2 any) int, offset, limit int) (*model.EditionListResult, error) {
	items, total, err := e.editionStore.ListEditions(filter, orderBy, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list editions from store: %w", err)
	}
	if err := e.addSubjectCategories(items); err != nil {
		return nil, err
	}
	return &model.EditionListResult{Items: items, Total: total, Offset: offset, Limit: limit}, nil
}

func (e *Edition) addSubjectCategories(editions []*model.Edition) error {
	keys := lo.Map(editions, func(edition *model.Edition, _ int) string { return edition.Key })
	valuesByEdition, err := e.featureResultStore.ListEditionFeatureValues(subjectCategoriesFeatureID, keys)
	if err != nil {
		return fmt.Errorf("failed to load edition subject categories: %w", err)
	}
	for _, edition := range editions {
		edition.SubjectCategories = parseSubjectCategories(valuesByEdition[edition.Key])
	}
	return nil
}

func parseSubjectCategories(values []string) []model.EditionSubjectCategory {
	categories := make([]model.EditionSubjectCategory, 0, len(values))
	for _, value := range values {
		category, classification, found := strings.Cut(value, "::")
		if !found {
			continue
		}
		categories = append(categories, model.EditionSubjectCategory{
			Category:       category,
			Classification: classification,
		})
	}
	return categories
}

func (e *Edition) ListAllEditions() ([]*model.Edition, error) {
	const pageSize = 5000
	all := make([]*model.Edition, 0, pageSize)
	for offset := 0; ; offset += pageSize {
		page, err := e.ListEditions(nil, nil, offset, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Items...)
		if len(page.Items) < pageSize || len(all) >= page.Total {
			return all, nil
		}
	}
}

func (e *Edition) CreateEdition(ed *model.Edition, login string) (*model.Edition, error) {
	if ed.Key == "" {
		ed.Key = idgen.GenerateID("ed")
	}
	existing, err := e.GetEditionByID(ed.Key)
	if err != nil && !errors.Is(err, ErrEditionNotFound) {
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
		return nil, fmt.Errorf("%w: edition with key %s does not exist", ErrEditionNotFound, m.Key)
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
	if ed == nil {
		return nil, fmt.Errorf("%w: edition with key %s does not exist", ErrEditionNotFound, key)
	}
	if err := e.addSubjectCategories([]*model.Edition{ed}); err != nil {
		return nil, err
	}
	return ed, nil
}

func (e *Edition) DeleteEdition(key string) error {
	// Check if edition exists
	existing, err := e.GetEditionByID(key)
	if err != nil {
		return fmt.Errorf("failed to check for existing edition: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("%w: edition with key %s does not exist", ErrEditionNotFound, key)
	}
	// Delete associated facsimiles
	facs, err := e.facsimileStore.ListFacsimiles([]string{key})
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
