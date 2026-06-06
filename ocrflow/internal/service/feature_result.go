package service

import (
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
)

type Result struct {
	store           *fpstore.FeatureResultSql
	annotationStore *fpstore.AnnotationSQL
	featureSvc      *Feature
	featurePropSvc  *FeatureProperty
}

func NewResult(store *fpstore.FeatureResultSql, annotationStore *fpstore.AnnotationSQL, featureSvc *Feature, featurePropSvc *FeatureProperty) *Result {
	return &Result{store: store, annotationStore: annotationStore, featureSvc: featureSvc, featurePropSvc: featurePropSvc}
}

func (r *Result) ListResults(scope feature.ExecScope, keys []string, features []string, fallbackToOrigin bool) (res []*feature.Result, err error) {
	if fallbackToOrigin && scope.Type == feature.ScopeTypeDataset && len(keys) == 0 {
		ann, err := r.annotationStore.GetAnnotation(scope.DatasetID, scope.AnnotationID)
		if err != nil {
			return nil, err
		}
		if ann != nil && strings.TrimSpace(ann.Pages) != "" {
			keys, err = pagesparser.Range(ann.Pages)
			if err != nil {
				return nil, fmt.Errorf("parse annotation pages: %w", err)
			}
		}
	}
	res, err = r.store.List(scope, keys, features, fallbackToOrigin)
	if err != nil {
		return nil, err
	}
	feats, err := r.featureSvc.ListFeatures(scope.DefScope, nil)
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
		return strings.Compare(a.Key, b.Key)
	})
	return res, nil
}

func (r *Result) CreateResult(results []*feature.Result, pushToOrigin bool) ([]*feature.Result, error) {
	for _, result := range results {
		if err := r.store.Create(result, pushToOrigin); err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (r *Result) CreateResults(results []*feature.Result, pushToOrigin bool) error {
	return r.store.CreateBatch(results, pushToOrigin)
}

func (r *Result) ListResultsForExecutionPolicy(exec *feature.Execution, features []string) ([]*feature.Result, error) {
	pushToOrigin := exec.Scope.Type != feature.ScopeTypeEditions && exec.Policy != nil && exec.Policy.PushToOrigin
	return r.store.ListForExecutionPolicy(exec.Scope, exec.Keys, features, pushToOrigin)
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

func (r *Result) CopyResults(datasetID, srcAnnID, dstDatasetID, dstAnnID string) error {
	return r.store.CopyResults(datasetID, srcAnnID, dstDatasetID, dstAnnID)
}
