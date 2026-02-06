package featureplat

import (
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/samber/lo"
)

type Feature struct {
	m map[string]*featureplat.Feature
}

func NewFeature() *Feature {
	mockMap := make(map[string]*featureplat.Feature)
	mockMap["fea_ts8621"] = &featureplat.Feature{
		Meta: ocrflow.Meta{
			ID:        "fea_ts8621",
			Name:      "institutions",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		IsRoot:    true,
		IsDefault: true,
	}
	mockMap["fea_ts8622"] = &featureplat.Feature{
		Meta: ocrflow.Meta{
			ID:        "fea_ts8622",
			Name:      "normalized institutions",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		IsRoot:    false,
		IsDefault: true,
	}

	return &Feature{
		m: mockMap,
	}
}

func (f *Feature) ListFeatures(expandOptions []featureplat.FeatureExpandOptions) ([]*featureplat.Feature, error) {
	fs := lo.Values(f.m)
	// for _, opt := range expandOptions {...}
	return fs, nil
}

func (f *Feature) CreateFeature(m *featureplat.Feature) (*featureplat.Feature, error) {
	m.ID = idgen.GenerateID("fea")
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	f.m[m.ID] = m
	return m, nil
}

func (f *Feature) Delete(collectionId, id string, force bool) error {
	if _, err := f.GetFeature(collectionId, id, nil); err != nil {
		return err
	}
	delete(f.m, id)
	return nil
}

func (f *Feature) GetFeature(collectionId, id string, expandOptions []featureplat.FeatureExpandOptions) (*featureplat.Feature, error) {
	feat, exists := f.m[id]
	if !exists {
		return nil, fmt.Errorf("feature with id %s not found", id)
	}
	return feat, nil
}

func (f *Feature) UpdateFeature(collectionId, id string, updated *featureplat.Feature) (*featureplat.Feature, error) {
	existing, err := f.GetFeature(collectionId, id, nil)
	if err != nil {
		return nil, err
	}
	existing.Name = updated.Name
	existing.IsDefault = updated.IsDefault
	existing.UpdatedAt = time.Now()
	return existing, nil
}
