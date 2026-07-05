package feature

import (
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

type Execution struct {
	common.Meta
	Scope ExecScope `json:"scope"`
	// Keys is optional, if not provided, the execution will run on all keys of the dataset.
	Keys         []string             `json:"keys,omitempty"`
	Apply        []ExecutionApplyItem `json:"apply"`
	Policy       *ExecutionPolicy     `json:"policy,omitempty"`
	Status       ExecutionStatus      `json:"status"`
	StatusReason string               `json:"status_reason"`
}

type ExecutionStatus string

const (
	ExecutionStatusSuccess    ExecutionStatus = "success"
	ExecutionStatusFailed     ExecutionStatus = "failed"
	ExecutionStatusInProgress ExecutionStatus = "in_progress"
	ExecutionStatusCanceling  ExecutionStatus = "canceling"
	ExecutionStatusCanceled   ExecutionStatus = "canceled"
)

func ToExecutionsStatus(s string) ExecutionStatus {
	switch s {
	case string(ExecutionStatusSuccess):
		return ExecutionStatusSuccess
	case string(ExecutionStatusFailed):
		return ExecutionStatusFailed
	case string(ExecutionStatusInProgress):
		return ExecutionStatusInProgress
	case string(ExecutionStatusCanceling):
		return ExecutionStatusCanceling
	case string(ExecutionStatusCanceled):
		return ExecutionStatusCanceled
	default:
		return ""
	}
}

type ExecutionApplyItem struct {
	Feature  string `json:"feature"`
	Revision string `json:"revision"`
}

type ExecutionPolicy struct {
	SkipIf       []ExecutionSkipIf `json:"skip_if"`
	PushToOrigin bool              `json:"push_to_origin"`
}

type ExecutionSkipIf string

const (
	ExecutionSkipIfFeatureExist  ExecutionSkipIf = "feature_exist"
	ExecutionSkipIfRevisionExist ExecutionSkipIf = "revision_exist"
	ExecutionSkipIfHumanReviewed ExecutionSkipIf = "human_reviewed"
	ExecutionSkipIfValueNotEmpty ExecutionSkipIf = "value_not_empty"
)
