package featureplat

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store/featureplat"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

type Feature struct {
	store *fpstore.FeatureSQL
}

// NewFeature returns a new Feature service using the given store (e.g. *storefeatureplat.FeatureSQL).
func NewFeature(store *fpstore.FeatureSQL) *Feature {
	return &Feature{store: store}
}

func (f *Feature) ListFeatures(collectionID string, expandOptions []featureplat.FeatureExpandOptions) ([]*featureplat.Feature, error) {
	fs, err := f.store.List(collectionID)
	if err != nil {
		return nil, err
	}
	// TODO: apply expandOptions (latest_revision, revisions) when revision store is available
	_ = expandOptions
	return fs, nil
}

func (f *Feature) CreateFeature(collectionID string, m *featureplat.Feature) (*featureplat.Feature, error) {
	m.CollectionID = collectionID
	m.ID = idgen.GenerateID("fea")
	if err := f.store.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (f *Feature) Delete(collectionId, id string, force bool) error {
	if _, err := f.GetFeature(collectionId, id, nil); err != nil {
		return err
	}
	return f.store.Delete(collectionId, id)
}

func (f *Feature) GetFeature(collectionId, id string, expandOptions []featureplat.FeatureExpandOptions) (*featureplat.Feature, error) {
	feat, err := f.store.GetByID(collectionId, id)
	if err != nil {
		return nil, err
	}
	_ = expandOptions
	return feat, nil
}

func (f *Feature) UpdateFeature(collectionId, id string, updated *featureplat.Feature) (*featureplat.Feature, error) {
	existing, err := f.store.GetByID(collectionId, id)
	if err != nil {
		return nil, err
	}
	existing.Name = updated.Name
	existing.Description = updated.Description
	existing.IsDefault = updated.IsDefault
	if err := f.store.Update(collectionId, id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}
