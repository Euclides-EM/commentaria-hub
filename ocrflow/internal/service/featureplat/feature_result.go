package featureplat

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store/featureplat"
)

type Result struct {
	store *fpstore.FeatureResultSQL
}

// NewResult returns a new Result service using the given store (e.g. *storefeatureplat.FeatureResultSQL).
func NewResult(store *fpstore.FeatureResultSQL) *Result {
	return &Result{store: store}
}

func (r *Result) ListResults(collectionID string, keys []string, features []string) ([]*featureplat.FeatureResult, error) {
	return r.store.List(collectionID, keys, features)
}

func (r *Result) CreateResult(m *featureplat.FeatureResult) (*featureplat.FeatureResult, error) {
	if err := r.store.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}
