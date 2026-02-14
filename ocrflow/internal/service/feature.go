package service

import (
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	store2 "github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

type Feature struct {
	store         *store2.FeatureSQL
	revisionStore *store2.FeatureRevisionSQL
}

func NewFeature(store *store2.FeatureSQL, revisionStore *store2.FeatureRevisionSQL) *Feature {
	return &Feature{store: store, revisionStore: revisionStore}
}

func (f *Feature) ListFeatures(datasetID string, expandOptions []feature.ExpandOptions) ([]*feature.Feature, error) {
	fs, err := f.store.List(datasetID)
	if err != nil {
		return nil, err
	}
	for _, feat := range fs {
		if err := f.applyExpand(feat, expandOptions); err != nil {
			return nil, err
		}
	}
	return fs, nil
}

func (f *Feature) CreateFeature(datasetID string, m *feature.Feature) (*feature.Feature, error) {
	if m.Color == "" {
		return nil, fmt.Errorf("feature color is required")
	}
	m.DatasetID = datasetID
	m.ID = idgen.GenerateID("fea")
	if err := f.store.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (f *Feature) Delete(datasetID, id string, force bool) error {
	if _, err := f.GetFeature(datasetID, id, nil); err != nil {
		return err
	}
	return f.store.Delete(datasetID, id)
}

func (f *Feature) GetFeature(datasetID, id string, expandOptions []feature.ExpandOptions) (*feature.Feature, error) {
	feat, err := f.store.GetByID(datasetID, id)
	if err != nil {
		return nil, err
	}
	if err := f.applyExpand(feat, expandOptions); err != nil {
		return nil, err
	}
	return feat, nil
}

func (f *Feature) applyExpand(feat *feature.Feature, expandOptions []feature.ExpandOptions) error {
	for _, opt := range expandOptions {
		switch opt {
		case feature.ExpandRevisions:
			revisions, err := f.revisionStore.ListByFeatureID(feat.DatasetID, feat.ID)
			if err != nil {
				return err
			}
			feat.Revisions = revisions
		case feature.ExpandLatestRevision:
			revisions, err := f.revisionStore.ListByFeatureID(feat.DatasetID, feat.ID)
			if err != nil {
				return err
			}
			if len(revisions) > 0 {
				feat.LatestRevision = revisions[0]
			}
		}
	}
	return nil
}

func (f *Feature) UpdateFeature(datasetID, id string, updated *feature.Feature) (*feature.Feature, error) {
	if updated.Color == "" {
		return nil, fmt.Errorf("feature color is required")
	}
	existing, err := f.store.GetByID(datasetID, id)
	if err != nil {
		return nil, err
	}
	existing.Name = updated.Name
	existing.Description = updated.Description
	existing.IsDefault = updated.IsDefault
	existing.Color = updated.Color
	if err := f.store.Update(datasetID, id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}
