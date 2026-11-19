package tei

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
	"github.com/samber/lo"
)

// buildStandOff populates doc.Header.StandOff from entities (Items or Profiles). Any entry with
// ObjectRef set is emitted as a relation (or interp fallback). Fact IDs are auto-generated.
func buildStandOff(pageOrKey string, entities []EntityItem, blocks []alto.TextBlock, tags alto.Tags) *model.StandOff {
	zoneCategories := zoneCategoriesFromBlocks(blocks, tags)
	if len(entities) == 0 && len(blocks) == 0 && len(zoneCategories) == 0 {
		return nil
	}
	so := &model.StandOff{}
	if len(entities) > 0 {
		so.InterpGrps = []model.InterpGrp{
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
		}
		so.SpanGrp = &model.SpanGrp{
			XmlID: SpanGrpHighlightsID,
			Type:  SpanGrpHighlightsType,
			Spans: buildHighlightSpans(entities),
		}
	}
	if len(zoneCategories) > 0 {
		so.InterpGrps = append(so.InterpGrps, model.InterpGrp{
			XmlID:   InterpGrpAltoCategoriesID,
			Type:    InterpGrpAltoCategoriesType,
			Interps: buildAltoCategoryInterps(zoneCategories),
		})
	}
	if len(blocks) > 0 {
		so.Certainties = lineCertaintiesFromBlocks(pageOrKey, blocks)
	}
	return so
}

func lineCertaintiesFromBlocks(pageOrKey string, blocks []alto.TextBlock) []model.Certainty {
	var certs []model.Certainty
	for bi, block := range blocks {
		for li, line := range block.Lines {
			if len(line.Strings) == 0 {
				continue
			}
			lID := lineID(pageOrKey, bi+1, li+1)
			if len(line.Strings) > 1 {
				log.Printf("warning: line %s has more than one string, skipping certainty annotation", lID)
			}
			if line.Strings[0].WC == 0 {
				continue
			}
			certs = append(certs, model.Certainty{
				Target: "#" + lID,
				Locus:  "value",
				Degree: fmt.Sprintf("%.3f", line.Strings[0].WC),
			})
		}
	}
	return certs
}

func buildHighlightCategoriesInterps(entities []EntityItem) []model.Interp {
	cats := lo.Uniq(lo.Map(entities, func(e EntityItem, _ int) string {
		return e.Category
	}))

	sort.Strings(cats)

	return lo.Map(cats, func(cat string, _ int) model.Interp {
		return model.Interp{
			XmlID: interpCategoryID(cat),
			Text:  cat,
		}
	})
}

func buildAltoCategoryInterps(cats []string) []model.Interp {
	sort.Strings(cats)
	return lo.Map(cats, func(cat string, _ int) model.Interp {
		return model.Interp{
			XmlID: interpAltoCategoryID(cat),
			Text:  cat,
		}
	})
}

func buildHighlightPropsInterps(entities []EntityItem) []model.Interp {
	var allProps []string
	for _, entity := range entities {
		for propKey := range entity.Properties {
			allProps = append(allProps, propKey)
		}
	}

	allProps = lo.Uniq(allProps)
	sort.Strings(allProps)

	return lo.Map(allProps, func(prop string, _ int) model.Interp {
		return model.Interp{
			XmlID: interpPropID(prop),
			Text:  prop,
		}
	})
}

func altoCategoryAna(tagRefs string, tags alto.Tags) string {
	labelsByID := altoTagLabelsByID(tags)
	var refs []string
	for _, tagRef := range strings.Fields(tagRefs) {
		label, ok := labelsByID[tagRef]
		if !ok {
			continue
		}
		category := normalizeAltoZoneCategory(label)
		if category == "" {
			continue
		}
		refs = append(refs, "#"+interpAltoCategoryID(category))
	}
	return strings.Join(lo.Uniq(refs), " ")
}

func zoneCategoriesFromBlocks(blocks []alto.TextBlock, tags alto.Tags) []string {
	labelsByID := altoTagLabelsByID(tags)
	var categories []string
	for _, block := range blocks {
		for _, tagRef := range strings.Fields(block.TagRefs) {
			label, ok := labelsByID[tagRef]
			if !ok {
				continue
			}
			category := normalizeAltoZoneCategory(label)
			if category != "" {
				categories = append(categories, category)
			}
		}
	}
	return lo.Uniq(categories)
}

func altoTagLabelsByID(tags alto.Tags) map[string]string {
	labelsByID := make(map[string]string, len(tags.OtherTags))
	for _, tag := range tags.OtherTags {
		if tag.ID != "" && tag.Label != "" {
			labelsByID[tag.ID] = tag.Label
		}
	}
	return labelsByID
}

func normalizeAltoZoneCategory(label string) string {
	label = strings.TrimSpace(label)
	label = strings.TrimSuffix(label, "Zone")
	label = strings.ReplaceAll(label, "Zone-", "-")

	var b strings.Builder
	var lastUnderscore bool
	var prev rune
	for i, r := range label {
		isUpperBoundary := i > 0 && unicode.IsUpper(r) && (unicode.IsLower(prev) || unicode.IsDigit(prev))
		if isUpperBoundary && !lastUnderscore {
			b.WriteRune('_')
			lastUnderscore = true
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
		} else if !lastUnderscore && b.Len() > 0 {
			b.WriteRune('_')
			lastUnderscore = true
		}
		prev = r
	}

	return strings.Trim(b.String(), "_")
}

func buildHighlightSpans(entities []EntityItem) []model.Span {
	return lo.Map(entities, func(e EntityItem, i int) model.Span {
		propKeys := lo.Keys(e.Properties)
		sort.Strings(propKeys)
		notes := make([]model.Note, 0, len(propKeys))
		for _, propKey := range propKeys {
			notes = append(notes, model.Note{
				Ana:  "#" + interpPropID(propKey),
				Text: e.Properties[propKey],
			})
		}
		return model.Span{
			XmlID: spanHighlightID(i),
			From:  "#" + startMentionAnchorID(i),
			To:    "#" + endMentionAnchorID(i),
			Ana:   "#" + interpCategoryID(e.Category),
			Notes: notes,
		}
	})
}
