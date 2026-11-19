package store

import (
	"errors"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/cache"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/csv"
	"github.com/samber/lo"
)

type GeoCSV struct {
	itemsMetadataDir string
	cacheStore       *cache.Cache
}

func NewGeoCSV(itemsMetadataDir string) *GeoCSV {
	return &GeoCSV{
		itemsMetadataDir: itemsMetadataDir,
		cacheStore:       cache.NewCache(),
	}
}

const (
	relCities = "cities.csv"
)

func (s *GeoCSV) WarmCache() error {
	return s.cacheStore.Warmup(func() (map[string]interface{}, error) {
		m, err := s.LoadAllCities()
		if err != nil {
			return nil, err
		}
		return lo.MapValues(m, func(ed *model.City, _ string) any {
			return ed
		}), nil
	})
}

func (s *GeoCSV) LoadAllCities() (map[string]*model.City, error) {
	_, rows, err := csv.LoadCSVRecords(path.Join(s.itemsMetadataDir, relCities))
	if err != nil {
		return nil, err
	}

	cities := make(map[string]*model.City)
	for _, row := range rows {
		name := strings.TrimSpace(row["city"])
		if name == "" {
			continue
		}
		lat, err := strconv.ParseFloat(strings.TrimSpace(row["lat"]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude for city %q: %w", name, err)
		}
		lon, err := strconv.ParseFloat(strings.TrimSpace(row["lon"]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude for city %q: %w", name, err)
		}
		cities[name] = &model.City{
			Name:      name,
			Longitude: lon,
			Latitude:  lat,
		}
	}

	return cities, nil
}

func (s *GeoCSV) ListCities() ([]*model.City, error) {
	if !s.cacheStore.IsWarm() {
		return nil, errors.New("not warm cache")
	}

	_, cities, total, err := s.cacheStore.GetBulk(
		func(_ any) bool { return true },
		func(k1 string, k2 string, v1 any, v2 any) int {
			return strings.Compare(k1, k2)
		}, 0, 10000)
	if total > 10000 {
		log.Printf("ListCities: total cities in cache exceeds limit: %d", total)
	}
	return lo.Map(cities, func(c any, _ int) *model.City {
		if city, ok := c.(*model.City); ok {
			return city
		}
		log.Printf("ListCities: unexpected type in cache: %T", c)
		return nil
	}), err
}
