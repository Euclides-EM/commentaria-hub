package alto

import (
	"sort"
	"strings"
)

// LimitCategoryZones keeps at most maxCount zones of category per page.
// When more exist, keeps the ones closest to keepPosition ("top", "bottom", "left", "right")
// and removes the rest from the ALTO.
func LimitCategoryZones(a *Alto, category string, maxCount int, keepPosition string) error {
	if a == nil || category == "" || maxCount < 0 {
		return nil
	}
	if maxCount == 0 {
		// Keep zero → remove all of that category (same as remove_categories for that category)
		return ApplyRemoveCategoriesALTO(a, []string{category})
	}

	idToLabel := make(map[string]string, len(a.Tags.OtherTags))
	for _, ot := range a.Tags.OtherTags {
		idToLabel[ot.ID] = ot.Label
	}

	keepPosition = strings.TrimSpace(strings.ToLower(keepPosition))

	for pi := range a.Layout.Page {
		page := &a.Layout.Page[pi]
		blocks := page.PrintSpace.TextBlocks
		if len(blocks) == 0 {
			continue
		}

		// Pairs of (block index, sort key) for blocks that have this category
		type indexed struct {
			idx int
			key float64
		}
		var withCategory []indexed
		for i := range blocks {
			tb := &blocks[i]
			if !blockHasLabel(tb, idToLabel, category) {
				continue
			}
			var key float64
			switch keepPosition {
			case "top":
				key = tb.VPOS
			case "bottom":
				key = -(tb.VPOS + tb.Height) // negate so "largest bottom" sorts first
			case "left":
				key = tb.HPOS
			case "right":
				key = -(tb.HPOS + tb.Width)
			default:
				key = tb.VPOS // default like "top"
			}
			withCategory = append(withCategory, indexed{i, key})
		}

		if len(withCategory) <= maxCount {
			continue
		}

		// Sort by key ascending (bottom/right use negative so they sort correctly)
		sort.Slice(withCategory, func(i, j int) bool {
			return withCategory[i].key < withCategory[j].key
		})

		// Indices to remove = those beyond the first maxCount
		toRemove := make(map[int]struct{})
		for _, p := range withCategory[maxCount:] {
			toRemove[p.idx] = struct{}{}
		}

		kept := blocks[:0]
		for i := range blocks {
			if _, remove := toRemove[i]; remove {
				continue
			}
			kept = append(kept, blocks[i])
		}
		page.PrintSpace.TextBlocks = kept
	}
	return nil
}

func blockHasLabel(tb *TextBlock, idToLabel map[string]string, label string) bool {
	for _, id := range strings.Fields(tb.TagRefs) {
		if idToLabel[id] == label {
			return true
		}
	}
	return false
}
