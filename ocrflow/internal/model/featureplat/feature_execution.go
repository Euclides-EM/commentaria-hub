package featureplat

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
)

type FeatureExecution struct {
	ocrflow.Meta
	Collection string                      `json:"collection"`
	Keys       []string                    `json:"keys,omitempty"` // if not provided -> run on all keys.
	Apply      []FeatureExecutionApplyItem `json:"apply"`
	Policy     *FeatureExecutionPolicy     `json:"policy,omitempty"`
	Status     FeatureExecutionStatus      `json:"status"` // "success"|"failed"|"in_prpgress"|"canceling"|"canceled"
}

type FeatureExecutionStatus string

const (
	FeatureExecutionStatusSuccess    FeatureExecutionStatus = "success"
	FeatureExecutionStatusFailed     FeatureExecutionStatus = "failed"
	FeatureExecutionStatusInProgress FeatureExecutionStatus = "in_progress"
	FeatureExecutionStatusCanceling  FeatureExecutionStatus = "canceling"
	FeatureExecutionStatusCanceled   FeatureExecutionStatus = "canceled"
)

func ToFeatureExecutionsStatus(s string) FeatureExecutionStatus {
	switch s {
	case string(FeatureExecutionStatusSuccess):
		return FeatureExecutionStatusSuccess
	case string(FeatureExecutionStatusFailed):
		return FeatureExecutionStatusFailed
	case string(FeatureExecutionStatusInProgress):
		return FeatureExecutionStatusInProgress
	case string(FeatureExecutionStatusCanceling):
		return FeatureExecutionStatusCanceling
	case string(FeatureExecutionStatusCanceled):
		return FeatureExecutionStatusCanceled
	default:
		return ""
	}
}

type FeatureExecutionApplyItem struct {
	Feature  string `json:"feature"`
	Revision string `json:"revision"`
}

type FeatureExecutionPolicy struct {
	SkipIf FeatureExecutionSkipIf `json:"skip_if"`
}

type FeatureExecutionSkipIf string // "feature_exist","revision_exist","human_reviewed"

const (
	FeatureExecutionSkipIfFeatureExist  FeatureExecutionSkipIf = "feature_exist"
	FeatureExecutionSkipIfRevisionExist FeatureExecutionSkipIf = "revision_exist"
	FeatureExecutionSkipIfHumanReviewed FeatureExecutionSkipIf = "human_reviewed"
)
