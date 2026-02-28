package service

import (
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store"
)

type Result struct {
	store          *fpstore.FeatureResultSql
	featureSvc     *Feature
	featurePropSvc *FeatureProperty
}

func NewResult(store *fpstore.FeatureResultSql, featureSvc *Feature, featurePropSvc *FeatureProperty) *Result {
	return &Result{store: store, featureSvc: featureSvc, featurePropSvc: featurePropSvc}
}

func (r *Result) ListResults(datasetID, annotationID string, keys []string, features []string) ([]*feature.Result, error) {
	res, err := r.store.List(datasetID, annotationID, keys, features)
	if err != nil {
		return nil, err
	}
	feats, err := r.featureSvc.ListFeatures(datasetID, nil)
	if err != nil {
		return nil, err
	}
	featureByID := lo.SliceToMap(feats, func(f *feature.Feature) (string, *feature.Feature) {
		return f.ID, f
	})

	for _, result := range res {
		if err := r.enrichWithDynamicProperties(result, featureByID[result.FeatureID]); err != nil {
			return nil, err
		}
	}

	slices.SortFunc(res, func(a, b *feature.Result) int {
		return strings.Compare(a.PageKey, b.PageKey)
	})
	return res, nil
}

func (r *Result) CreateResult(results []*feature.Result) ([]*feature.Result, error) {
	for _, result := range results {
		if err := r.store.Create(result); err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (r *Result) CreateResults(results []*feature.Result) error {
	return r.store.CreateBatch(results)
}

func (r *Result) enrichWithDynamicProperties(result *feature.Result, feat *feature.Feature) error {
	if feat == nil {
		return fmt.Errorf("feature %s not found", result.FeatureID)
	}
	for _, propKey := range feat.Properties {
		for i := range result.Values {
			if result.Values[i].Properties == nil {
				result.Values[i].Properties = make(map[string]string)
			}
			if _, exists := result.Values[i].Properties[propKey]; exists {
				continue
			}
			propVal, err := r.featurePropSvc.CalcValByPropertyKey(result.Values[i].Surface, propKey)
			if err != nil {
				return fmt.Errorf("failed to calculate feature property %s: %v", propKey, err)
			}
			result.Values[i].Properties[propKey] = propVal
		}
	}
	return nil
}
