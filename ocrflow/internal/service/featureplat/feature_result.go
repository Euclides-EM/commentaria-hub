package featureplat

import (
	"slices"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
)

type Result struct {
	s []*featureplat.FeatureResult
}

func NewResult() *Result {
	return &Result{
		s: []*featureplat.FeatureResult{},
	}
}

func (r *Result) ListResults(keys []string, features []string) ([]*featureplat.FeatureResult, error) {
	var result []*featureplat.FeatureResult
	for _, res := range r.s {
		if (len(keys) == 0 || slices.Contains(keys, res.Key)) && (len(features) == 0 || slices.Contains(features, res.Feature)) {
			result = append(result, res)
		}
	}
	return result, nil
}

func (r *Result) CreateResult(m *featureplat.FeatureResult) (*featureplat.FeatureResult, error) {
	r.s = append(r.s, m)
	return m, nil
}
