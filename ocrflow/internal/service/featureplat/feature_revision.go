package featureplat

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

// FeatureRevisionStore is the minimal store interface used by the feature revision service.
type FeatureRevisionStore interface {
	ListByFeatureID(collectionID, featureID string) ([]*featureplat.FeatureRevision, error)
	GetByID(collectionID, featureID, revisionID string) (*featureplat.FeatureRevision, error)
	Create(collectionID, featureID string, rev *featureplat.FeatureRevision) error
	Update(collectionID, featureID, revisionID string, rev *featureplat.FeatureRevision) error
	Delete(collectionID, featureID, revisionID string) error
}

type Revision struct {
	store FeatureRevisionStore
}

// NewRevision returns a new Revision service using the given store (e.g. *storefeatureplat.FeatureRevisionSQL).
func NewRevision(store FeatureRevisionStore) *Revision {
	return &Revision{store: store}
}

func (fr *Revision) ListFeatureRevisions(collectionId, featureId string) ([]*featureplat.FeatureRevision, error) {
	return fr.store.ListByFeatureID(collectionId, featureId)
}

func (fr *Revision) CreateFeatureRevision(collectionId, featureId string, m *featureplat.FeatureRevision) (*featureplat.FeatureRevision, error) {
	m.CollectionID = collectionId
	m.FeatureID = featureId
	m.ID = idgen.GenerateID("rev")
	if err := fr.store.Create(collectionId, featureId, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (fr *Revision) GetFeatureRevision(collectionId, featureId, revisionId string) (*featureplat.FeatureRevision, error) {
	return fr.store.GetByID(collectionId, featureId, revisionId)
}

func (fr *Revision) DeleteFeatureRevision(collectionId, featureId, revisionId string) error {
	return fr.store.Delete(collectionId, featureId, revisionId)
}

func (fr *Revision) UpdateFeatureRevision(collectionId, featureId, revisionId string, m *featureplat.FeatureRevision) (*featureplat.FeatureRevision, error) {
	existing, err := fr.store.GetByID(collectionId, featureId, revisionId)
	if err != nil {
		return nil, err
	}
	existing.Name = m.Name
	existing.Description = m.Description
	existing.Prompt = m.Prompt
	existing.Regex = m.Regex
	existing.ExecutionStrategy = m.ExecutionStrategy
	existing.Note = m.Note
	existing.Type = m.Type
	existing.Features = m.Features
	if err := fr.store.Update(collectionId, featureId, revisionId, existing); err != nil {
		return nil, err
	}
	return existing, nil
}
