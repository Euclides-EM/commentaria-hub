package model

type OCRModelType string

const (
	OCRModelTypeSegment OCRModelType = "segment"
	OCRModelTypeOCR     OCRModelType = "text"
)

type Model struct {
	Meta      `json:",inline"`
	LocalPath string       `json:"local_path" readonly:"true"`
	Type      OCRModelType `json:"type"`
	RunWith   string       `json:"run_with"`
	Name      string       `json:"name"`
}

func (m *Model) DeepCopy() *Model {
	if m == nil {
		return nil
	}
	return &Model{
		Meta:      m.Meta.DeepCopy(),
		LocalPath: m.LocalPath,
		Type:      m.Type,
		RunWith:   m.RunWith,
		Name:      m.Name,
	}
}
