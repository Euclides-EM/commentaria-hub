package service

import (
	"errors"
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/samber/lo"
)

type Revision struct {
	store             *fpstore.FeatureRevisionSQL
	featureProperties *FeatureProperty
}

// NewRevision returns a new Revision service using the given store (e.g. *storefeatureplat.FeatureRevisionSQL).
func NewRevision(store *fpstore.FeatureRevisionSQL, featureProperties *FeatureProperty) *Revision {
	return &Revision{store: store, featureProperties: featureProperties}
}

func (fr *Revision) ListFeatureRevisions(featureId string) ([]*feature.Revision, error) {
	return fr.store.ListByFeatureID(featureId)
}

func (fr *Revision) ListFeatureRevisionsInScope(scope feature.DefScope, featureId string) ([]*feature.Revision, error) {
	return fr.store.ListByFeatureIDInScope(scope, featureId)
}

func (fr *Revision) CreateFeatureRevision(featureId string, m *feature.Revision) (*feature.Revision, error) {
	if err := fr.validate(m); err != nil {
		return nil, err
	}
	m.FeatureID = featureId
	m.ID = idgen.GenerateID("rev")
	if err := fr.store.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (fr *Revision) GetFeatureRevision(featureId, revisionId string) (*feature.Revision, error) {
	return fr.store.GetByID(featureId, revisionId)
}

func (fr *Revision) GetFeatureRevisionInScope(scope feature.DefScope, featureId, revisionId string) (*feature.Revision, error) {
	return fr.store.GetByIDInScope(scope, featureId, revisionId)
}

func (fr *Revision) validate(m *feature.Revision) error {
	if m.Prompt == "" && m.Categorizer == "" {
		return errors.New("either prompt or categorizer is required")
	}
	if m.Prompt != "" && m.Categorizer != "" {
		return errors.New("cannot have both prompt and categorizer")
	}
	if m.Categorizer != "" {
		properties := fr.featureProperties.ListFeaturePropertyKeys()
		if !lo.Contains(properties, m.Categorizer) {
			return fmt.Errorf("categorizer %q is not a valid feature property key", m.Categorizer)
		}
	}
	return nil
}
