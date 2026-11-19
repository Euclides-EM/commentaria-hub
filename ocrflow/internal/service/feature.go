package service

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/feature"
	store2 "github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/idgen"
)

type Feature struct {
	store           *store2.FeatureSQL
	revisionStore   *store2.FeatureRevisionSQL
	featureProperty *FeatureProperty
}

func NewFeature(store *store2.FeatureSQL, revisionStore *store2.FeatureRevisionSQL, featureProperty *FeatureProperty) *Feature {
	return &Feature{store: store, revisionStore: revisionStore, featureProperty: featureProperty}
}

func (f *Feature) ListFeatures(scope feature.DefScope, expandOptions []feature.ExpandOptions) ([]*feature.Feature, error) {
	fs, err := f.store.ListFeatures(scope)
	if err != nil {
		return nil, err
	}
	for _, feat := range fs {
		if err := f.applyExpand(feat, expandOptions); err != nil {
			return nil, err
		}
		feat.Properties = lo.Uniq(append(feat.Properties, f.featureProperty.ListDefaultFeaturePropertyKeys()...))
	}
	return fs, nil
}

func (f *Feature) Create(m *feature.Feature) (*feature.Feature, error) {
	if err := f.validate(m); err != nil {
		return nil, err
	}
	m.ID = idgen.GenerateID("fea")
	if err := f.store.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (f *Feature) DeleteFeature(id string, force bool) error {
	feat, err := f.GetFeature(id, nil)
	if err != nil {
		return err
	}
	return f.store.DeleteFeature(feat.Scope, id)
}

func (f *Feature) DeleteFeatureInScope(scope feature.DefScope, id string, force bool) error {
	feat, err := f.GetFeatureInScope(scope, id, nil)
	if err != nil {
		return err
	}
	return f.store.DeleteFeature(feat.Scope, id)
}

func (f *Feature) DeleteFeaturesInScope(scope feature.DefScope, ids []string, force bool) ([]string, error) {
	deleted := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if err := f.DeleteFeatureInScope(scope, id, force); err != nil {
			return deleted, err
		}
		deleted = append(deleted, id)
	}
	return deleted, nil
}

func (f *Feature) GetFeature(id string, expandOptions []feature.ExpandOptions) (*feature.Feature, error) {
	feat, err := f.store.GetFeatureByID(id)
	if err != nil {
		return nil, err
	}
	if err := f.applyExpand(feat, expandOptions); err != nil {
		return nil, err
	}
	feat.Properties = lo.Uniq(append(feat.Properties, f.featureProperty.ListDefaultFeaturePropertyKeys()...))
	return feat, nil
}

func (f *Feature) GetFeatureInScope(scope feature.DefScope, id string, expandOptions []feature.ExpandOptions) (*feature.Feature, error) {
	feat, err := f.store.GetFeatureByIDInScope(scope, id)
	if err != nil {
		return nil, err
	}
	if err := f.applyExpand(feat, expandOptions); err != nil {
		return nil, err
	}
	feat.Properties = lo.Uniq(append(feat.Properties, f.featureProperty.ListDefaultFeaturePropertyKeys()...))
	return feat, nil
}

func (f *Feature) applyExpand(feat *feature.Feature, expandOptions []feature.ExpandOptions) error {
	for _, opt := range expandOptions {
		switch opt {
		case feature.ExpandRevisions:
			revisions, err := f.revisionStore.ListByFeatureIDInScope(feat.Scope, feat.ID)
			if err != nil {
				return err
			}
			feat.Revisions = revisions
		case feature.ExpandLatestRevision:
			revisions, err := f.revisionStore.ListByFeatureIDInScope(feat.Scope, feat.ID)
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

func (f *Feature) UpdateFeature(updated *feature.Feature) (*feature.Feature, error) {
	err := f.validate(updated)
	if err != nil {
		return nil, err
	}
	existing, err := f.store.GetFeatureByIDInScope(updated.Scope, updated.ID)
	if err != nil {
		return nil, err
	}
	existing.Name = updated.Name
	existing.Description = updated.Description
	existing.IsDefault = updated.IsDefault
	existing.Color = updated.Color
	existing.Properties = updated.Properties
	if err := f.store.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (f *Feature) validate(feat *feature.Feature) error {
	if feat.Name == "" {
		return fmt.Errorf("feature name is required")
	}
	if feat.Color == "" {
		return fmt.Errorf("feature color is required")
	}
	properties := f.featureProperty.ListFeaturePropertyKeys()
	for _, key := range feat.Properties {
		if !lo.Contains(properties, key) {
			return fmt.Errorf("invalid feature property key: %s", key)
		}
	}
	return nil
}
