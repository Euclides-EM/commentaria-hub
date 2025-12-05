package model

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
)

type Annotation struct {
	Meta        `json:",inline"`
	Description string                     `json:"description"`
	Pages       string                     `json:"pages"`
	AltoDir     string                     `json:"alto_dir" readonly:"true"`
	YoloDir     string                     `json:"yolo_dir" readonly:"true"`
	RoboflowDir string                     `json:"roboflow_dir" readonly:"true"`
	DatasetID   string                     `json:"dataset_id" readonly:"true"`
	ApplyRules  *annotationrule.ApplyRules `json:"apply_rules" readonly:"true"`
}
