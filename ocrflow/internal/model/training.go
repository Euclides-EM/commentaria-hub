package model

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type TrainingStatus string

const (
	TrainingStatusRunning   TrainingStatus = "running"
	TrainingStatusCompleted TrainingStatus = "completed"
	TrainingStatusFailed    TrainingStatus = "failed"
)

type Training struct {
	common.Meta    `json:",inline"`
	OriginModel    *Model                  `json:"origin_model"`
	AnnotationSets []*annotation.Reference `json:"annotation_sets"`
	Status         TrainingStatus          `json:"status"`
}
