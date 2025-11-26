package service

import (
	"errors"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"path"
)

type Model struct {
	m         map[string]*model.Model
	modelsDir string
}

func NewModelService(modelsDir string) *Model {
	return &Model{
		m: map[string]*model.Model{
			"CapricciosaM": {
				Meta:      model.NewMeta("CapricciosaM"),
				LocalPath: path.Join(modelsDir, "CapricciosaM.pt"),
				Type:      model.OCRModelTypeSegment,
			},
			"Gallicorpor": {
				Meta:      model.NewMeta("Gallicorpor"),
				LocalPath: path.Join(modelsDir, "Gallicorpor.mlmodel"),
				Type:      model.OCRModelTypeOCR,
			},
		},
	}
}

func (m *Model) List() ([]*model.Model, error) {
	modelsList := make([]*model.Model, 0, len(m.m))
	for _, retrieved := range m.m {
		modelsList = append(modelsList, retrieved.DeepCopy())
	}
	return modelsList, nil
}

func (m *Model) Get(id string) (*model.Model, error) {
	retrieved, ok := m.m[id]
	if !ok {
		return nil, errors.New("model not found")
	}
	return retrieved.DeepCopy(), nil
}
