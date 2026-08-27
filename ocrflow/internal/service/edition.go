package service

import (
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/idgen"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/search"
	"github.com/samber/lo"
)

// todo: add interfaces to all services

var ErrEditionNotFound = errors.New("edition not found")

const subjectCategoriesFeatureID = "m_classifier"
const editionFeatureFilterPrefix = "feature:"
const shelfmarkPropertiesFilterField = "shelfmarkProperties"

const (
	shelfmarkAvailable             = "shelfmark_available"
	facsimileAvailable             = "facsimile_available"
	copyrightStatusUnknown         = "copyright_status_unknown"
	externalTranscriptionAvailable = "external_transcription_available"
	internalTranscriptionAvailable = "internal_transcription_available"
)

type editionFeatureFilter struct {
	matchingEditionIDs map[string]struct{}
	include            bool
}

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

// SearchEditions applies API search filters, including edition feature-backed fields.
func (e *Edition) SearchEditions(query search.Query, offset, limit int) (*model.EditionListResult, error) {
	query.FieldsFilter = maps.Clone(query.FieldsFilter)
	query.FilterIncludes = maps.Clone(query.FilterIncludes)
	shelfmarkPropertyValues := query.FieldsFilter[shelfmarkPropertiesFilterField]
	shelfmarkPropertiesInclude := true
	if configured, exists := query.FilterIncludes[shelfmarkPropertiesFilterField]; exists {
		shelfmarkPropertiesInclude = configured
	}
	delete(query.FieldsFilter, shelfmarkPropertiesFilterField)
	delete(query.FilterIncludes, shelfmarkPropertiesFilterField)
	features := make(map[string][]string)
	featureFilters := make([]editionFeatureFilter, 0)
	for field, allowed := range query.FieldsFilter {
		if !strings.HasPrefix(field, editionFeatureFilterPrefix) || len(allowed) == 0 {
			continue
		}
		featureID := strings.TrimPrefix(field, editionFeatureFilterPrefix)
		if featureID == "" {
			continue
		}

		features[featureID] = allowed

		matchingIDs, err := e.featureResultStore.ListEditionIDsByFeatureValues(featureID, allowed)
		if err != nil {
			return nil, fmt.Errorf("filter editions by feature %q: %w", featureID, err)
		}
		include := true
		if configured, exists := query.FilterIncludes[field]; exists {
			include = configured
		}
		featureFilters = append(featureFilters, editionFeatureFilter{
			matchingEditionIDs: matchingIDs,
			include:            include,
		})
		delete(query.FieldsFilter, field)
		delete(query.FilterIncludes, field)
	}

	baseFilter := query.FilterFunc()
	filter := func(value any) bool {
		if !baseFilter(value) {
			return false
		}
		edition, ok := value.(*model.Edition)
		if !ok {
			return false
		}
		if len(shelfmarkPropertyValues) > 0 && editionHasAnyShelfmarkProperty(edition, shelfmarkPropertyValues) != shelfmarkPropertiesInclude {
			return false
		}
		for _, featureFilter := range featureFilters {
			_, matched := featureFilter.matchingEditionIDs[edition.Key]
			if matched != featureFilter.include {
				return false
			}
		}
		return true
	}

	items, total, err := e.editionStore.ListEditions(filter, query.OrderByFunc(), offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list editions from store: %w", err)
	}
	if err := e.addSubjectCategories(items); err != nil {
		return nil, err
	}
	return &model.EditionListResult{Items: items, Total: total, Offset: offset, Limit: limit}, nil
}

func editionHasAnyShelfmarkProperty(edition *model.Edition, allowed []string) bool {
	for _, shelfmark := range edition.Shelfmarks {
		for _, property := range allowed {
			switch property {
			case shelfmarkAvailable:
				if strings.TrimSpace(shelfmark.Shelfmark) != "" {
					return true
				}
			case facsimileAvailable:
				if strings.TrimSpace(shelfmark.Scan) != "" {
					return true
				}
			case copyrightStatusUnknown:
				if strings.TrimSpace(shelfmark.Copyright) == "" {
					return true
				}
			case externalTranscriptionAvailable:
				if shelfmark.TranscriptionAvailable == model.EditionShelfmarkTranscriptionExternal {
					return true
				}
			case internalTranscriptionAvailable:
				if shelfmark.TranscriptionAvailable == model.EditionShelfmarkTranscriptionInternal {
					return true
				}
			}
		}
	}
	return false
}

func (e *Edition) ListEditions(filter func(e any) bool, orderBy func(e1, e2 any) int, offset, limit int) (*model.EditionListResult, error) {
	items, total, err := e.editionStore.ListEditions(filter, orderBy, offset, limit)
	if err != nil {
		return nil, err
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
