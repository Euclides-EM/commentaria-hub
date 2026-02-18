package service

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

type Revision struct {
	store *fpstore.FeatureRevisionSQL
}

// NewRevision returns a new Revision service using the given store (e.g. *storefeatureplat.FeatureRevisionSQL).
func NewRevision(store *fpstore.FeatureRevisionSQL) *Revision {
	return &Revision{store: store}
}

func (fr *Revision) ListFeatureRevisions(datasetID, featureId string) ([]*feature.Revision, error) {
	return fr.store.ListByFeatureID(datasetID, featureId)
}

func (fr *Revision) CreateFeatureRevision(datasetID, featureId string, m *feature.Revision) (*feature.Revision, error) {
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
	existing, err := fr.store.GetByID(datasetID, featureId, revisionId)
	if err != nil {
		return nil, err
	}
	existing.Name = m.Name
	existing.Description = m.Description
	existing.Prompt = m.Prompt
	existing.Regex = m.Regex
	existing.ExecutionStrategy = m.ExecutionStrategy
	existing.Note = m.Note
	if err := fr.store.Update(datasetID, featureId, revisionId, existing); err != nil {
		return nil, err
	}
	return existing, nil
}
