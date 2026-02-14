package feature

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type Execution struct {
	common.Meta
	DatasetID    string `json:"dataset_id"`
	AnnotationID string `json:"annotation_id"`
	// Keys is optional, if not provided, the execution will run on all keys of the dataset.
	Keys   []string             `json:"keys,omitempty"`
	Apply  []ExecutionApplyItem `json:"apply"`
	Policy *ExecutionPolicy     `json:"policy,omitempty"`
	Status ExecutionStatus      `json:"status"`
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
	DatasetID    string `json:"dataset_id"`
	AnnotationId string `json:"annotation_id"`
	Feature      string `json:"feature"`
	Revision     string `json:"revision"`
}

type ExecutionPolicy struct {
	SkipIf ExecutionSkipIf `json:"skip_if"`
}

type ExecutionSkipIf string

const (
	ExecutionSkipIfFeatureExist  ExecutionSkipIf = "feature_exist"
	ExecutionSkipIfRevisionExist ExecutionSkipIf = "revision_exist"
	ExecutionSkipIfHumanReviewed ExecutionSkipIf = "human_reviewed"
)
