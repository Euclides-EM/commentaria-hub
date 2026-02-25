package tei

import (
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
)

// buildEncodingDesc builds encodingDesc with a taxonomy of feature and fact categories from entities.
func buildEncodingDesc(entities []EntityItem) *model.EncodingDesc {
	featNames := make(map[string]bool)
	featAnas := make(map[string]bool) // ana values e.g. #feat_person -> category feat_person
	factTypes := make(map[string]bool)
	for _, it := range entities {
		if strings.TrimSpace(it.Type) == "" {
			// Collect Ana from mention items for taxonomy categories (e.g. #feat_person -> Person).
			if ana := strings.TrimSpace(it.Ana); ana != "" && strings.HasPrefix(ana, "#") {
				cid := strings.TrimPrefix(ana, "#")
				if cid != "" {
					featAnas[cid] = true
				}
			}
			continue
		}
		if it.Type == "feature_name" && strings.TrimSpace(it.Value) != "" {
			featNames[it.Value] = true
		}
	}
	var names []string
	for n := range featNames {
		names = append(names, n)
	}
	sort.Strings(names)
	var anaIDs []string
	for cid := range featAnas {
		anaIDs = append(anaIDs, cid)
	}
	sort.Strings(anaIDs)
	var types []string
	for t := range factTypes {
		types = append(types, t)
	}
	sort.Strings(types)

	var categories []model.Category
	for _, cid := range anaIDs {
		categories = append(categories, model.Category{
			XmlID:   cid,
			CatDesc: categoryIDToDesc(cid),
		})
	}
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

// categoryIDToDesc returns a human-readable description for a category xml:id (e.g. feat_person -> Person).
func categoryIDToDesc(cid string) string {
	cid = strings.TrimPrefix(cid, "feat_")
	cid = strings.ReplaceAll(cid, "_", " ")
	cid = strings.TrimSpace(cid)
	if cid == "" {
		return cid
	}
	return strings.ToUpper(cid[:1]) + cid[1:]
}
