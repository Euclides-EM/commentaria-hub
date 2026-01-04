package alto

import (
	"strings"
)

func ExtractCategoryContents(a *Alto, category string, lineBreakSeperator string) ([]string, error) {
	catTagIDs := findTagIDsByLabel(a, category)
	if len(catTagIDs) == 0 {
		return nil, nil
	}

	var contents []string
	for _, page := range a.Layout.Page {
		for _, block := range page.PrintSpace.TextBlocks {
			if !tagrefsContainsAny(block.TagRefs, catTagIDs) {
				continue
			}
			h := buildContentFromBlock(block, lineBreakSeperator)
			if h != "" {
				contents = append(contents, h)
			}
		}
	}

	return contents, nil
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
