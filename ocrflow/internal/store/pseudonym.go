package store

import (
	"errors"
	"log"
	"path"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/cache"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/csv"
	"github.com/samber/lo"
)

type PseudonymCSV struct {
	itemsMetadataDir string
	cacheStore       *cache.Cache
}

func NewPseudonymCSV(itemsMetadataDir string) *PseudonymCSV {
	return &PseudonymCSV{
		itemsMetadataDir: itemsMetadataDir,
		cacheStore:       cache.NewCache(),
	}
}

const (
	relPseudonyms       = "pseudonyms.csv"
	pseudonymsCacheLimit = 100000
)

func (s *PseudonymCSV) WarmCache() error {
	return s.cacheStore.Warmup(func() (map[string]interface{}, error) {
		pseudonyms, err := s.LoadAllPseudonyms()
		if err != nil {
			return nil, err
		}
		return lo.SliceToMap(pseudonyms, func(p *model.Pseudonym) (string, any) {
			return pseudonymCacheKey(p), p
		}), nil
	})
}

func (s *PseudonymCSV) LoadAllPseudonyms() ([]*model.Pseudonym, error) {
	_, rows, err := csv.LoadCSVRecords(path.Join(s.itemsMetadataDir, relPseudonyms))
	if err != nil {
		return nil, err
	}

	pseudonyms := make([]*model.Pseudonym, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row["name"])
		pseudonym := strings.TrimSpace(row["pseudonym"])
		position := strings.TrimSpace(row["position"])
		source := strings.TrimSpace(row["source"])
		if name == "" && pseudonym == "" && position == "" && source == "" {
			continue
		}

		pseudonyms = append(pseudonyms, &model.Pseudonym{
			Name:      name,
			Pseudonym: pseudonym,
			Position:  position,
			Source:    source,
		})
	}

	return pseudonyms, nil
}

func (s *PseudonymCSV) ListPseudonyms() ([]*model.Pseudonym, error) {
	if !s.cacheStore.IsWarm() {
		return nil, errors.New("not warm cache")
	}

	_, values, total, err := s.cacheStore.GetBulk(
		func(_ any) bool { return true },
		func(k1 string, k2 string, _ any, _ any) int {
			return strings.Compare(k1, k2)
		}, 0, pseudonymsCacheLimit)
	if total > pseudonymsCacheLimit {
		log.Printf("ListPseudonyms: total pseudonyms in cache exceeds limit: %d", total)
	}

	return lo.Map(values, func(v any, _ int) *model.Pseudonym {
		if pseudonym, ok := v.(*model.Pseudonym); ok {
			return pseudonym
		}
		log.Printf("ListPseudonyms: unexpected type in cache: %T", v)
		return nil
	}), err
}

func pseudonymCacheKey(p *model.Pseudonym) string {
	return strings.Join([]string{p.Name, p.Pseudonym, p.Position, p.Source}, "\x00")
}
