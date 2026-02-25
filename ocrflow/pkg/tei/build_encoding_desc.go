package tei

import (
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
)

// buildEncodingDesc builds encodingDesc with a taxonomy of feature and fact categories from entities.
func buildEncodingDesc(entities []EntityItem) *model.EncodingDesc {
	featNames := make(map[string]bool)
	factTypes := make(map[string]bool)
	for _, it := range entities {
		if strings.TrimSpace(it.Type) == "" {
			continue
		}
		if it.Type == "feature_name" && strings.TrimSpace(it.Value) != "" {
			featNames[it.Value] = true
		}
		if strings.TrimSpace(it.ObjectRef) != "" {
			factTypes[it.Type] = true
		}
	}
	var names []string
	for n := range featNames {
		names = append(names, n)
	}
	sort.Strings(names)
	var types []string
	for t := range factTypes {
		types = append(types, t)
	}
	sort.Strings(types)

	var categories []model.Category
	for _, n := range names {
		categories = append(categories, model.Category{
			XmlID:   featureNameToCategoryID(n),
			CatDesc: n,
		})
	}
	for _, t := range types {
		cid := factTypeToCategoryID(t)
		categories = append(categories, model.Category{
			XmlID:   cid,
			CatDesc: t + " fact",
		})
	}
	// Always include fact type for feature-assignment relations (standOff listRelation).
	categories = append(categories, model.Category{
		XmlID:   "fact_feature_assignment",
		CatDesc: "Feature assignment (fact)",
	})
	if len(categories) == 0 {
		return nil
	}
	return &model.EncodingDesc{
		ClassDecl: &model.ClassDecl{
			Taxonomy: &model.Taxonomy{
				XmlID:      "features",
				Categories: categories,
			},
		},
	}
}
