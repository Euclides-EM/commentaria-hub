package alto

import (
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/geo"
)

// ResolveOverlapWithPriority removes zones of suppressedCategory that overlap
// zones of dominantCategory by at least minOverlapPct percent of the suppressed
// zone's area. Removed blocks are deleted from the ALTO.
func ResolveOverlapWithPriority(a *Alto, dominantCategory, suppressedCategory string, minOverlapRatio float64) error {
	if a == nil || dominantCategory == "" || suppressedCategory == "" {
		return nil
	}
	if minOverlapRatio <= 0 {
		minOverlapRatio = 0
	}
	if minOverlapRatio > 1 {
		minOverlapRatio = 1
	}

	dominantIDs := labelToIDSet(a, dominantCategory)
	suppressedIDs := labelToIDSet(a, suppressedCategory)
	if len(dominantIDs) == 0 || len(suppressedIDs) == 0 {
		return nil
	}

	for pi := range a.Layout.Page {
		page := &a.Layout.Page[pi]
		blocks := page.PrintSpace.TextBlocks
		if len(blocks) == 0 {
			continue
		}

		var dominantRects []geo.Rectangle
		var suppressedIndices []int
		for i, b := range blocks {
			if hasAnyTagRef(b.TagRefs, dominantIDs) {
				dominantRects = append(dominantRects, *geo.RectangleFromLeftBottomCorner(b.HPOS, b.VPOS, b.Width, b.Height))
			}
			if hasAnyTagRef(b.TagRefs, suppressedIDs) {
				suppressedIndices = append(suppressedIndices, i)
			}
		}

		toRemove := make(map[int]struct{})
		for _, si := range suppressedIndices {
			sb := blocks[si]
			suppRect := geo.RectangleFromLeftBottomCorner(sb.HPOS, sb.VPOS, sb.Width, sb.Height)
			for dr := range dominantRects {
				ratio := suppRect.OverlapRatio(&dominantRects[dr])
				if ratio >= minOverlapRatio {
					toRemove[si] = struct{}{}
					break
				}
			}
		}

		if len(toRemove) == 0 {
			continue
		}
		var newBlocks []TextBlock
		for i, b := range blocks {
			if _, remove := toRemove[i]; !remove {
				newBlocks = append(newBlocks, b)
			}
		}
		page.PrintSpace.TextBlocks = newBlocks
	}
	return nil
}

func labelToIDSet(a *Alto, label string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, ot := range a.Tags.OtherTags {
		if ot.Label == label {
			out[ot.ID] = struct{}{}
		}
	}
	return out
}
