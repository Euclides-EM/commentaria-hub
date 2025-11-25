package service

import (
	"errors"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
)

type Model struct {
	m map[string]*model.Model
}

func NewModelService() *Model {
	return &Model{
		m: map[string]*model.Model{
			"1": {
				Meta:      model.NewMeta("1"),
				KrakenRef: "10.5281/zenodo.10592716",
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
