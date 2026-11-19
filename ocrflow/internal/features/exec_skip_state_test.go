package features

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/feature"
)

func TestExecutionSkipStateValueNotEmpty(t *testing.T) {
	state := NewExecutionSkipState()
	policy := &feature.ExecutionPolicy{
		SkipIf: []feature.ExecutionSkipIf{
			feature.ExecutionSkipIfFeatureExist,
			feature.ExecutionSkipIfValueNotEmpty,
		},
	}

	state.Add(&feature.Result{
		FeatureID: "language",
		Key:       "page-1",
		Values: []feature.ResultValue{
			{Surface: "  "},
		},
	})

	reasons := state.SkipReasons(policy, "page-1", "language", "rev-x")
	if len(reasons) != 1 || reasons[0] != feature.ExecutionSkipIfFeatureExist {
		t.Fatalf("expected only feature_exist for whitespace value, got %#v", reasons)
	}

	state.Add(&feature.Result{
		FeatureID: "language",
		Key:       "page-2",
		Values: []feature.ResultValue{
			{Surface: "Spanish"},
		},
	})

	reasons = state.SkipReasons(policy, "page-2", "language", "rev-y")
	if len(reasons) != 2 || reasons[0] != feature.ExecutionSkipIfFeatureExist || reasons[1] != feature.ExecutionSkipIfValueNotEmpty {
		t.Fatalf("expected feature_exist and value_not_empty, got %#v", reasons)
	}
}
