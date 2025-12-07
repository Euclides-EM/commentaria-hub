package alto

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/annotationrules"
	"math"
	"strings"
)

// ApplyAddMarginALTO adds a margin on the given side(s) of all TextBlocks
// whose category (resolved from TAGREFS -> OtherTag.LABEL) matches cfg.Category.
//
// Margin is in pixels. Positive values expand the block outward.
func ApplyAddMarginALTO(doc *Alto, cfg *annotationrules.AddMargin) error {
	if doc == nil || cfg == nil {
		return nil
	}
	if cfg.Margin == 0 {
		return nil
	}

	// Build ID -> Label map from <OtherTag>
	idToLabel := make(map[string]string, len(doc.Tags.OtherTags))
	for _, ot := range doc.Tags.OtherTags {
		idToLabel[ot.ID] = ot.Label
	}

	// Decide which sides to modify
	resolveSides := func(s annotationrules.Side) []annotationrules.Side {
		switch s {
		case annotationrules.SideAll:
			return []annotationrules.Side{annotationrules.SideLeft, annotationrules.SideRight, annotationrules.SideTop, annotationrules.SideBottom}
		case annotationrules.SideHorizontal:
			return []annotationrules.Side{annotationrules.SideLeft, annotationrules.SideRight}
		case annotationrules.SideVertical:
			return []annotationrules.Side{annotationrules.SideTop, annotationrules.SideBottom}
		default:
			return []annotationrules.Side{s}
		}
	}

	for pi := range doc.Layout.Page {
		page := &doc.Layout.Page[pi]
		ps := &page.PrintSpace

		for i := range ps.TextBlocks {
			tb := &ps.TextBlocks[i]
			if tb == nil || tb.Width <= 0 || tb.Height <= 0 {
				continue
			}

			// Check if this block belongs to the requested category
			tagIDs := strings.Fields(tb.TagRefs)
			matchesCategory := false
			for _, id := range tagIDs {
				if idToLabel[id] == cfg.Category {
					matchesCategory = true
					break
				}
			}
			if !matchesCategory {
				continue
			}

			// Work in float64 then round back
			x := float64(tb.HPOS)
			y := float64(tb.VPOS)
			w := float64(tb.Width)
			h := float64(tb.Height)

			if w <= 0 || h <= 0 {
				continue
			}

			sides := resolveSides(cfg.Side)
			margin := cfg.Margin

			for _, side := range sides {
				switch side {
				case annotationrules.SideLeft:
					x -= margin
					w += margin
				case annotationrules.SideRight:
					w += margin
				case annotationrules.SideTop:
					y -= margin
					h += margin
				case annotationrules.SideBottom:
					h += margin
				default:
					return fmt.Errorf("unsupported side %q on page %s", side, page.ID)
				}
			}

			if w <= 0 || h <= 0 {
				// Do not apply if geometry becomes invalid
				continue
			}

			// Commit back to ints
			tb.HPOS = int(math.Round(x))
			tb.VPOS = int(math.Round(y))
			tb.Width = int(math.Round(w))
			tb.Height = int(math.Round(h))

			// Keep polygon in sync with the rectangle
			updateTextBlockPolygon(tb)
		}
	}

	return nil
}
