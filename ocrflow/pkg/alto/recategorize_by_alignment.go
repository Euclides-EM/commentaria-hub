package alto

import (
	"strings"
)

// intervalsShareLine reports whether there exists a line (perpendicular to the axis)
// that passes through both intervals. So for "horizontal" we use Y-intervals: they
// must overlap (or be within tolerancePx). For "vertical" we use X-intervals.
// overlap = min(aHigh, bHigh) - max(aLow, bLow); aligned if overlap >= -tolerancePx.
func intervalsShareLine(aLow, aHigh, bLow, bHigh, tolerancePx float64) bool {
	overlap := min(aHigh, bHigh) - max(aLow, bLow)
	return overlap >= -tolerancePx
}

// RecategorizeByAlignment changes zones with originalCategory to targetCategory when
// there is some horizontal or vertical line that passes through both the candidate zone
// and at least one zone of relativeCategory (i.e. their Y-ranges or X-ranges overlap,
// or are within tolerancePx). No area overlap is required.
func RecategorizeByAlignment(a *Alto, originalCategory, targetCategory, relativeCategory, alignment string, tolerancePx float64) error {
	if a == nil || originalCategory == "" || targetCategory == "" || relativeCategory == "" {
		return nil
	}
	if tolerancePx < 0 {
		tolerancePx = 0
	}

	targetID, err := resolveTagID(a, targetCategory)
	if err != nil {
		return err
	}
	originalIDs := labelToIDSet(a, originalCategory)
	relativeIDs := labelToIDSet(a, relativeCategory)
	if len(originalIDs) == 0 || len(relativeIDs) == 0 {
		return nil
	}

	alignHorizontal := strings.TrimSpace(strings.ToLower(alignment)) == "horizontal"

	for pi := range a.Layout.Page {
		page := &a.Layout.Page[pi]
		blocks := page.PrintSpace.TextBlocks
		if len(blocks) == 0 {
			continue
		}

		// Reference intervals from relative_to.category blocks (same page)
		type interval struct{ low, high float64 }
		var refIntervals []interval
		for _, b := range blocks {
			if !hasAnyTagRef(b.TagRefs, relativeIDs) {
				continue
			}
			if alignHorizontal {
				refIntervals = append(refIntervals, interval{b.VPOS, b.VPOS + b.Height})
			} else {
				refIntervals = append(refIntervals, interval{b.HPOS, b.HPOS + b.Width})
			}
		}
		if len(refIntervals) == 0 {
			continue
		}

		// Recategorize original_category blocks for which some line passes through both zones
		for i := range blocks {
			b := &blocks[i]
			if !hasAnyTagRef(b.TagRefs, originalIDs) {
				continue
			}
			var candLow, candHigh float64
			if alignHorizontal {
				candLow, candHigh = b.VPOS, b.VPOS+b.Height
			} else {
				candLow, candHigh = b.HPOS, b.HPOS+b.Width
			}
			for _, ref := range refIntervals {
				if intervalsShareLine(candLow, candHigh, ref.low, ref.high, tolerancePx) {
					b.TagRefs = targetID
					break
				}
			}
		}
	}
	return nil
}
