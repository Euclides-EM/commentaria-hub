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

func (fr *Revision) ListFeatureRevisions(datasetID, featureId string) ([]*feature.Revision, error) {
	return fr.store.ListByFeatureID(datasetID, featureId)
}

func (fr *Revision) CreateFeatureRevision(datasetID, featureId string, m *feature.Revision) (*feature.Revision, error) {
	if err := fr.validate(m); err != nil {
		return nil, err
	}
	m.DatasetID = datasetID
	m.FeatureID = featureId
	m.ID = idgen.GenerateID("rev")
	if err := fr.store.Create(datasetID, featureId, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (fr *Revision) GetFeatureRevision(datasetID, featureId, revisionId string) (*feature.Revision, error) {
	return fr.store.GetByID(datasetID, featureId, revisionId)
}

func (fr *Revision) DeleteFeatureRevision(datasetID, featureId, revisionId string) error {
	return fr.store.Delete(datasetID, featureId, revisionId)
}

func (fr *Revision) UpdateFeatureRevision(datasetID, featureId, revisionId string, m *feature.Revision) (*feature.Revision, error) {
	if err := fr.validate(m); err != nil {
		return nil, err
	}
	existing, err := fr.store.GetByID(datasetID, featureId, revisionId)
	if err != nil {
		return nil, err
	}
	existing.Name = m.Name
	existing.Description = m.Description
	existing.Prompt = m.Prompt
	existing.Categorizer = m.Categorizer
	if err := fr.store.Update(datasetID, featureId, revisionId, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (fr *Revision) validate(m *feature.Revision) error {
	if m.Name == "" {
		return errors.New("name is required")
	}
	if m.Prompt == "" && m.Categorizer == "" {
		return errors.New("either prompt or categorizer is required")
	}
	if m.Prompt != "" && m.Categorizer != "" {
		return errors.New("cannot have both prompt and categorizer")
	}
	if m.Categorizer != "" {
		properties, err := fr.featureProperties.ListFeaturePropertyKeys()
		if err != nil {
			return err
		}
		if !lo.Contains(properties, m.Categorizer) {
			return fmt.Errorf("categorizer %q is not a valid feature property key", m.Categorizer)
		}
	}
	return nil
}
