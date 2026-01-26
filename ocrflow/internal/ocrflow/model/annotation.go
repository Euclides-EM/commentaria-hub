package model

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
)

type Annotation struct {
	Meta               `json:",inline"`
	Pages              string                         `json:"pages"`
	Segmented          bool                           `json:"segmented" readonly:"true"`
	GroundTruth        bool                           `json:"ground_truth"`
	Ocred              bool                           `json:"ocred" readonly:"true"`
	DatasetID          string                         `json:"dataset_id" readonly:"true"`
	AppliedRules       annotationrule.AnnotationRules `json:"applied_rules" readonly:"true"`
	OriginAnnotationID string                         `json:"origin_annotation_id,omitempty" readonly:"true"`
	PipelineStage      annotationrule.PipelineStage   `json:"pipeline_stage,omitempty" readonly:"true"`
}
