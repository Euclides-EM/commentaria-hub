package ocrflow

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type TrainingStatus string

const (
	TrainingStatusRunning   TrainingStatus = "running"
	TrainingStatusCompleted TrainingStatus = "completed"
	TrainingStatusFailed    TrainingStatus = "failed"
)

type Training struct {
	common.Meta    `json:",inline"`
	OriginModel    *Model                 `json:"origin_model"`
	AnnotationSets []*AnnotationReference `json:"annotation_sets"`
	Status         TrainingStatus         `json:"status"`
}

type AnnotationReference struct {
	ID        string `json:"id"`
	DatasetID string `json:"dataset_id"`
}
