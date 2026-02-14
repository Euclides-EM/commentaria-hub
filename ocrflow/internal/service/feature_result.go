package service

import (
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

func (r *Result) ListResults(datasetID string, keys []string, features []string) ([]*feature.Result, error) {
	return r.store.List(datasetID, keys, features)
}

func (r *Result) CreateResult(m *feature.Result) (*feature.Result, error) {
	if err := r.store.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}
