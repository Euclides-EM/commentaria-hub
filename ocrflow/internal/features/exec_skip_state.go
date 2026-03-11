package features

import (
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

type ExecutionSkipState struct {
	featureExists  map[string]struct{}
	revisionExists map[string]struct{}
	humanReviewed  map[string]struct{}
}

func NewExecutionSkipState() *ExecutionSkipState {
	return &ExecutionSkipState{
		featureExists:  make(map[string]struct{}),
		revisionExists: make(map[string]struct{}),
		humanReviewed:  make(map[string]struct{}),
	}
}

func (s *ExecutionSkipState) ShouldSkip(policy *feature.ExecutionPolicy, key, featureID, revisionID string) bool {
	if policy == nil || len(policy.SkipIf) == 0 {
		return false
	}

	featureKey := featureID + "::" + key
	revisionKey := featureKey + "::" + revisionID
	for _, rule := range policy.SkipIf {
		switch rule {
		case feature.ExecutionSkipIfFeatureExist:
			if _, ok := s.featureExists[featureKey]; ok {
				return true
			}
		case feature.ExecutionSkipIfRevisionExist:
			if _, ok := s.revisionExists[revisionKey]; ok {
				return true
			}
		case feature.ExecutionSkipIfHumanReviewed:
			if _, ok := s.humanReviewed[featureKey]; ok {
				return true
			}
		}
	}
	return false
}

func (s *ExecutionSkipState) Add(featureID, key, revisionID string, reviewed bool) {
	s.featureExists[featureID+"::"+key] = struct{}{}
	if strings.TrimSpace(revisionID) != "" {
		s.revisionExists[featureID+"::"+key+"::"+revisionID] = struct{}{}
	}
	if reviewed {
		s.humanReviewed[featureID+"::"+key] = struct{}{}
	}
}
