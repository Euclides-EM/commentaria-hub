package model

import "github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/common"

type OCRModelLocation string

const (
	OCRModelLocationLocal    OCRModelLocation = "local"
	OCRModelLocationRoboflow OCRModelLocation = "roboflow"
)

type OCRModelAlgorithmFamily string

const (
	OCRModelAlgorithmFamilyYOLO OCRModelAlgorithmFamily = "yolo"
)

type Model struct {
	Meta            `json:",inline"`
	Type            common.OCRModelType     `json:"type"`
	Location        OCRModelLocation        `json:"location"`
	AlgorithmFamily OCRModelAlgorithmFamily `json:"algorithm_family,omitempty"`
	// LocalPath is the path to the model file on the local filesystem. It is relevant only for local models.
	LocalPath         string                 `json:"local_path" readonly:"true"`
	Categories        []string               `json:"categories,omitempty"`
	BaseModelID       string                 `json:"base_model_id,omitempty"`
	BaseAnnotations   []*AnnotationReference `json:"base_annotations,omitempty"`
	UsedInAnnotations []*AnnotationReference `json:"used_in_annotations,omitempty"`
}
