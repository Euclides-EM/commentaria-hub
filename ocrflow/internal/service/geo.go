package service

import (
	"fmt"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
)

type Geo struct {
	geoStore *store.GeoCSV
}

func NewGeoService(geoStore *store.GeoCSV) *Geo {
	return &Geo{
		geoStore: geoStore,
	}
}

func (g *Geo) ListCities() ([]*model.City, error) {
	cities, err := g.geoStore.ListCities()
	if err != nil {
		return nil, fmt.Errorf("failed to list cities from store: %w", err)
	}
	return cities, nil
}
