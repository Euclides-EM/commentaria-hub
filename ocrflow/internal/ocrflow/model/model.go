package model

type OCRModelType string

const (
	OCRModelTypeSegment OCRModelType = "segment"
	OCRModelTypeOCR     OCRModelType = "text"
	OCRModelTypeUnknown OCRModelType = "unknown"
)

func OCRModelTypeFromExt(ext string) OCRModelType {
	switch ext {
	case ".pt":
		return OCRModelTypeSegment
	case ".mlmodel":
		return OCRModelTypeOCR
	default:
		return OCRModelTypeUnknown
	}
}

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
	Type            OCRModelType            `json:"type"`
	Location        OCRModelLocation        `json:"location"`
	AlgorithmFamily OCRModelAlgorithmFamily `json:"algorithm_family,omitempty"`
	// LocalPath is the path to the model file on the local filesystem. It is relevant only for local models.
	LocalPath       string                 `json:"local_path" readonly:"true"`
	Categories      []string               `json:"categories,omitempty"`
	BaseModelID     string                 `json:"base_model_id,omitempty"`
	BaseAnnotations []*AnnotationReference `json:"base_annotations,omitempty"`
}
