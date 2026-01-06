package model

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
)

type Annotation struct {
	Meta               `json:",inline"`
	Pages              string                          `json:"pages"`
	Segmented          bool                            `json:"segmented" readonly:"true"`
	DatasetID          string                          `json:"dataset_id" readonly:"true"`
	AppliedRules       []annotationrule.AnnotationRule `json:"applied_rules" readonly:"true"`
	OriginAnnotationID string                          `json:"origin_annotation_id,omitempty" readonly:"true"`
}
