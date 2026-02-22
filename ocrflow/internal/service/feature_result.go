package service

import (
	"slices"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store"
)

type Result struct {
	store *fpstore.FeatureResultSQL
}

// NewResult returns a new Result service using the given store (e.g. *storefeatureplat.FeatureResultSQL).
func NewResult(store *fpstore.FeatureResultSQL) *Result {
	return &Result{store: store}
}

func (r *Result) ListResults(datasetID, annotationID string, keys []string, features []string) ([]*feature.Result, error) {
	res, err := r.store.List(datasetID, annotationID, keys, features)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(res, func(a, b *feature.Result) int {
		return strings.Compare(a.Key, b.Key)
	})
	return res, nil
}

func (r *Result) CreateResult(m *feature.Result) (*feature.Result, error) {
	if err := r.store.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *Result) CreateResults(results []*feature.Result) error {
	return r.store.CreateBatch(results)
}
