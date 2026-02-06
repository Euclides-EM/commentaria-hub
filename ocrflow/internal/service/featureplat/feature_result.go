package featureplat

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
)

// FeatureResultStore is the minimal store interface used by the feature result service.
type FeatureResultStore interface {
	List(collectionID string, keys []string, features []string) ([]*featureplat.FeatureResult, error)
	Create(res *featureplat.FeatureResult) error
}

type Result struct {
	store FeatureResultStore
}

// NewResult returns a new Result service using the given store (e.g. *storefeatureplat.FeatureResultSQL).
func NewResult(store FeatureResultStore) *Result {
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
