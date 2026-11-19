package tei

import (
	"fmt"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

func CombineTEIsByKey(teis map[string]*model.TEI) (*model.CombinedTEI, error) {
	if len(teis) == 0 {
		return nil, fmt.Errorf("no TEIs provided")
	}

	items := make([]model.CombinedTEIItem, 0, len(teis))

	for key, tei := range teis {
		if tei == nil {
			return nil, fmt.Errorf("TEI for key %q is nil", key)
		}

		if tei.Xmlns == "" {
			tei.Xmlns = "http://www.tei-c.org/ns/1.0"
		}

		items = append(items, model.CombinedTEIItem{
			Key: key,
			TEI: tei,
		})
	}

	return &model.CombinedTEI{
		Xmlns: "http://example.com/api/tei-batch",
		Items: items,
	}, nil
}
