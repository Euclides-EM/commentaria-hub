package features

import (
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/feature"
)

type ExecutionSkipState struct {
	featureExists  map[string]struct{}
	revisionExists map[string]struct{}
	humanReviewed  map[string]struct{}
	valueExists    map[string]struct{}
}

func NewExecutionSkipState() *ExecutionSkipState {
	return &ExecutionSkipState{
		featureExists:  make(map[string]struct{}),
		revisionExists: make(map[string]struct{}),
		humanReviewed:  make(map[string]struct{}),
		valueExists:    make(map[string]struct{}),
	}
}

func (s *ExecutionSkipState) ShouldSkip(policy *feature.ExecutionPolicy, key, featureID, revisionID string) bool {
	return len(s.SkipReasons(policy, key, featureID, revisionID)) > 0
}

func (s *ExecutionSkipState) SkipReasons(policy *feature.ExecutionPolicy, key, featureID, revisionID string) []feature.ExecutionSkipIf {
	if policy == nil || len(policy.SkipIf) == 0 {
		return nil
	}

	featureKey := featureID + "::" + key
	revisionKey := featureKey + "::" + revisionID
	var reasons []feature.ExecutionSkipIf
	for _, rule := range policy.SkipIf {
		switch rule {
		case feature.ExecutionSkipIfFeatureExist:
			if _, ok := s.featureExists[featureKey]; ok {
				reasons = append(reasons, rule)
			}
		case feature.ExecutionSkipIfRevisionExist:
			if _, ok := s.revisionExists[revisionKey]; ok {
				reasons = append(reasons, rule)
			}
		case feature.ExecutionSkipIfHumanReviewed:
			if _, ok := s.humanReviewed[featureKey]; ok {
				reasons = append(reasons, rule)
			}
		case feature.ExecutionSkipIfValueNotEmpty:
			if _, ok := s.valueExists[featureKey]; ok {
				reasons = append(reasons, rule)
			}
		}
	}
	return reasons
}

func (s *ExecutionSkipState) Add(result *feature.Result) {
	if result == nil {
		return
	}

	featureKey := result.FeatureID + "::" + result.Key
	s.featureExists[featureKey] = struct{}{}
	if strings.TrimSpace(result.Source.Revision) != "" {
		s.revisionExists[featureKey+"::"+result.Source.Revision] = struct{}{}
	}
	if result.Source.Resp == "human" {
		s.humanReviewed[featureKey] = struct{}{}
	}
	for _, value := range result.Values {
		if strings.TrimSpace(value.Surface) != "" {
			s.valueExists[featureKey] = struct{}{}
			break
		}
	}
}
