package service

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/querylang"
)

type Dataset struct {
	m map[string]*models.Dataset
}

func (d *Dataset) ListDatasets(filter *querylang.Filter, sort querylang.Sort) ([]*models.Dataset, error) {
	return []*models.Dataset{}, nil
}

func NewDatasetService() *Dataset {
	return &Dataset{
		m: make(map[string]*models.Dataset),
	}
}
