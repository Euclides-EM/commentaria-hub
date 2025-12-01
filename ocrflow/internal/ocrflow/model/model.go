package model

type OCRModelType string

const (
	OCRModelTypeSegment OCRModelType = "segment"
	OCRModelTypeOCR     OCRModelType = "text"
)

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
	LocalPath       string                  `json:"local_path" readonly:"true"`
	Type            OCRModelType            `json:"type"`
	Location        OCRModelLocation        `json:"location"`
	AlgorithmFamily OCRModelAlgorithmFamily `json:"algorithm_family,omitempty"`
	Name            string                  `json:"name"`
	Categories      []string                `json:"categories,omitempty"`
}

func (m *Model) DeepCopy() *Model {
	if m == nil {
		return nil
	}
	return &Model{
		Meta:            m.Meta.DeepCopy(),
		LocalPath:       m.LocalPath,
		Type:            m.Type,
		Location:        m.Location,
		AlgorithmFamily: m.AlgorithmFamily,
		Name:            m.Name,
		Categories: func(src []string) []string {
			if src == nil {
				return nil
			}
			dst := make([]string, len(src))
			copy(dst, src)
			return dst
		}(m.Categories),
	}
}
