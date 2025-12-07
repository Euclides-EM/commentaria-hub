package alto

import (
	"strings"
)

// ApplyRemoveCategoriesALTO removes all TextBlocks whose TAGREFS
// resolve to *any* of the configured categories.
func ApplyRemoveCategoriesALTO(doc *Alto, categories []string) error {
	if doc == nil || len(categories) == 0 {
		return nil
	}

	// Convert slice to set for fast lookup
	catSet := make(map[string]struct{}, len(categories))
	for _, c := range categories {
		catSet[c] = struct{}{}
	}

	// Build ID -> LABEL map
	idToLabel := make(map[string]string, len(doc.Tags.OtherTags))
	for _, ot := range doc.Tags.OtherTags {
		idToLabel[ot.ID] = ot.Label
	}

	for pi := range doc.Layout.Page {
		page := &doc.Layout.Page[pi]
		ps := &page.PrintSpace

		if len(ps.TextBlocks) == 0 {
			continue
		}

		kept := ps.TextBlocks[:0] // reuse underlying array

		for i := range ps.TextBlocks {
			tb := &ps.TextBlocks[i]
			tagIDs := strings.Fields(tb.TagRefs)

			shouldRemove := false

			for _, id := range tagIDs {
				// Compare against OtherTag.LABEL
				label := idToLabel[id]
				if _, exists := catSet[label]; exists {
					shouldRemove = true
					break
				}
			}

			if shouldRemove {
				continue // drop this block
			}

			kept = append(kept, *tb)
		}

		ps.TextBlocks = kept
	}

	return nil
}
