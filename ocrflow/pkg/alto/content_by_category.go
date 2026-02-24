package alto

import (
	"strings"
)

type CategoryAndContent struct {
	Category           string
	Content            string
	VerticalPosition   float64
	HorizontalPosition float64
}

func NewCategoryAndContent(category, content string, vPos, hPos float64) *CategoryAndContent {
	return &CategoryAndContent{
		Category:           category,
		Content:            content,
		VerticalPosition:   vPos,
		HorizontalPosition: hPos,
	}
}

func ExtractCategoryContents(a *Alto, categories []string, lineBreakSeperator string) ([]*CategoryAndContent, error) {
	catToTagIDs := make(map[string]map[string]bool)
	for _, category := range categories {
		catToTagIDs[category] = findTagIDsByLabel(a, category)
	}

	var contents []*CategoryAndContent
	for _, page := range a.Layout.Page {
		for _, block := range page.PrintSpace.TextBlocks {
			for cat, catTagIDs := range catToTagIDs {
				if !tagrefsContainsAny(block.TagRefs, catTagIDs) {
					continue
				}
				h := buildContentFromBlock(block, lineBreakSeperator)
				if h != "" {
					contents = append(contents, NewCategoryAndContent(cat, h, block.VPOS, block.HPOS))
				}
			}
		}
	}

	return contents, nil
}

func ExtractTextContentFromBlock(b *TextBlock) string {
	lines := ExtractTextContentsFromBlock(b)
	combined := ""
	for _, content := range lines {
		c := strings.TrimSpace(content)
		if strings.HasSuffix(c, "¬") {
			combined += strings.TrimSuffix(c, "¬")
		} else {
			combined += c + " "
		}
	}
	combined = strings.TrimSpace(combined)
	return combined
}

func ExtractTextContentsFromBlock(b *TextBlock) []string {
	var contents []string
	for _, ln := range b.Lines {
		// A TextLine may contain multiple <String/> nodes. Join them with spaces.
		var chunks []string
		for _, s := range ln.Strings {
			txt := strings.TrimSpace(s.Content)
			if txt != "" {
				chunks = append(chunks, txt)
			}
		}
		lineText := strings.TrimSpace(strings.Join(chunks, " "))
		if lineText != "" {
			contents = append(contents, lineText)
		}
	}

	return contents
}

func ExtractBlocksByCategory(a *Alto, category string) ([]*TextBlock, error) {
	tagID := findTagIDsByLabel(a, category)
	var tbs []*TextBlock
	for _, page := range a.Layout.Page {
		for _, block := range page.PrintSpace.TextBlocks {
			if !tagrefsContainsAny(block.TagRefs, tagID) {
				continue
			}
			tbs = append(tbs, &block)
		}
	}

	return tbs, nil
}

func findTagIDsByLabel(a *Alto, label string) map[string]bool {
	ids := make(map[string]bool)
	for _, t := range a.Tags.OtherTags {
		if t.Label == label && t.ID != "" {
			ids[t.ID] = true
		}
	}
	return ids
}

func tagrefsContainsAny(tagrefs string, want map[string]bool) bool {
	if tagrefs == "" || len(want) == 0 {
		return false
	}
	for _, tok := range strings.Fields(tagrefs) {
		if want[tok] {
			return true
		}
	}
	return false
}

func buildContentFromBlock(b TextBlock, lineBreakSeperator string) string {
	var lineParts []string

	for _, ln := range b.Lines {
		// A TextLine may contain multiple <String/> nodes. Join them with spaces.
		var chunks []string
		for _, s := range ln.Strings {
			txt := strings.TrimSpace(s.Content)
			if txt != "" {
				chunks = append(chunks, txt)
			}
		}
		lineText := strings.TrimSpace(strings.Join(chunks, " "))
		if lineText != "" {
			lineParts = append(lineParts, lineText)
		}
	}

	return strings.Join(lineParts, lineBreakSeperator)
}
