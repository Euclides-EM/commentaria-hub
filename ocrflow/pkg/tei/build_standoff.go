package tei

import (
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
	"github.com/samber/lo"
)

// buildStandOff populates doc.Header.StandOff from entities (Items or Profiles). Any entry with
// ObjectRef set is emitted as a relation (or interp fallback). Fact IDs are auto-generated.
func buildStandOff(entities []EntityItem) *model.StandOff {
	if len(entities) == 0 {
		return nil
	}

	return &model.StandOff{
		InterpGrps: []model.InterpGrp{
			{
				XmlID:   InterpGrpCategoriesID,
				Type:    InterpGrpCategoriesType,
				Interps: buildHighlightCategoriesInterps(entities),
			},
			{
				XmlID:   InterpGrpPropsID,
				Type:    InterpGrpPropsType,
				Interps: buildHighlightPropsInterps(entities),
			},
		},
		SpanGrp: &model.SpanGrp{
			XmlID: SpanGrpHighlightsID,
			Type:  SpanGrpHighlightsType,
			Spans: buildHighlightSpans(entities),
		},
	}
}

func buildHighlightCategoriesInterps(entities []EntityItem) []model.Interp {
	cats := lo.Uniq(lo.Map(entities, func(e EntityItem, _ int) string {
		if e.Category != "" {
			return e.Category
		}
		return strings.TrimPrefix(e.Ana, "#")
	}))

	sort.Strings(cats)

	return lo.Map(cats, func(cat string, _ int) model.Interp {
		return model.Interp{
			XmlID: interpCategoryID(cat),
			Text:  cat,
		}
	})
}

func buildHighlightPropsInterps(entities []EntityItem) []model.Interp {
	var allProps []string
	for _, entity := range entities {
		for propKey, _ := range entity.Properties {
			allProps = append(allProps, propKey)
		}
	}

	allProps = lo.Uniq(allProps)

	return lo.Map(allProps, func(prop string, _ int) model.Interp {
		return model.Interp{
			XmlID: interpPropID(prop),
			Text:  prop,
		}
	})
}

func buildHighlightSpans(entities []EntityItem) []model.Span {
	// Only include mention-like entities (have location) so anchor indices match body
	mentions := lo.Filter(entities, func(e EntityItem, _ int) bool {
		return e.Start.LineID != ""
	})
	return lo.Map(mentions, func(e EntityItem, i int) model.Span {
		cat := e.Category
		if cat == "" {
			cat = strings.TrimPrefix(e.Ana, "#")
		}
		return model.Span{
			XmlID: spanHighlightID(i),
			From:  "#" + startMentionAnchorID(i),
			To:    "#" + endMentionAnchorID(i),
			Ana:   "#" + interpCategoryID(cat),
			Notes: lo.MapToSlice(e.Properties, func(propKey, propVal string) model.Note {
				return model.Note{
					Ana:  "#" + interpPropID(propKey),
					Text: propVal,
				}
			}),
		}
	})
}
