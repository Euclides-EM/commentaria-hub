package tei

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	transcriptionDivType = "transcription"
	translationDivType   = "translation"

	translationAnonBlockType   = "translation-block"
	transcriptionAnonBlockType = "transcription-block"

	textBlockZoneType = "text-block"
	textLineZoneType  = "text-line"

	InterpGrpCategoriesID   = "categories"
	InterpGrpCategoriesType = "highlight-categories"

	InterpGrpAltoCategoriesID   = "zone_categories"
	InterpGrpAltoCategoriesType = "zone_categories"

	InterpGrpPropsID   = "props"
	InterpGrpPropsType = "highlight-props"

	SpanGrpHighlightsID   = "highlights"
	SpanGrpHighlightsType = "highlights"
)

func surfaceID(pageID string) string {
	return "page_" + sanitizeID(pageID)
}

func transcriptionAnonBlockID(pageID string, blockIdx int) string {
	return "transcription_anon_blk_" + surfaceID(pageID) + "_" + strconv.Itoa(blockIdx)
}

func translationAnonBlockID(pageID string, lang string, blockIdx int) string {
	return "translation_anon_blk_" + surfaceID(pageID) + "_" + sanitizeID(lang) + "_" + strconv.Itoa(blockIdx)
}

func facZoneBlockID(pageID string, blockIdx int) string {
	return "zone_blk_" + surfaceID(pageID) + "_" + strconv.Itoa(blockIdx)
}

func facZoneLineID(pageID string, blockIdx, lineIdx int) string {
	return "z_line_" + surfaceID(pageID) + "_" + strconv.Itoa(blockIdx) + "_" + strconv.Itoa(lineIdx)
}

func lineID(pageID string, blockIdx, lineIdx int) string {
	return "line_" + surfaceID(pageID) + "_" + strconv.Itoa(blockIdx) + "_" + strconv.Itoa(lineIdx)
}

func startMentionAnchorID(mentionIdx int) string {
	return fmt.Sprintf("anchor_m%d_start", mentionIdx)
}

func endMentionAnchorID(mentionIdx int) string {
	return fmt.Sprintf("anchor_m%d_end", mentionIdx)
}

func interpCategoryID(cat string) string {
	return "cat_" + sanitizeID(cat)
}

func interpAltoCategoryID(cat string) string {
	return "zone_cat_" + sanitizeID(cat)
}

func interpPropID(prop string) string {
	return "prop_" + sanitizeID(prop)
}

func spanHighlightID(mentionIdx int) string {
	return "highlight_" + strconv.Itoa(mentionIdx)
}

func sanitizeID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "x"
	}
	// conservative: keep letters, digits, underscore, dash, replace others with underscore
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}
