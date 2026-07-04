package alto

import (
	"fmt"
	"math"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/annotationrules"
)

// ApplyStretchTowardsCategoryALTO applies StretchTowardsCategory logic to an ALTO document.
// It looks at TextBlocks whose TAGREFS resolve to StretchCategory and stretches them
// towards TextBlocks whose TAGREFS resolve to Towards, using the same rules as the COCO version.
func ApplyStretchTowardsCategoryALTO(doc *Alto, stc *annotationrules.StretchTowardsCategory) error {
	if doc == nil || stc == nil {
		return nil
	}

	// Build ID -> Label map from <OtherTag>
	idToLabel := make(map[string]string, len(doc.Tags.OtherTags))
	for _, ot := range doc.Tags.OtherTags {
		idToLabel[ot.ID] = ot.Label
	}

	// Walk pages
	for pi := range doc.Layout.Page {
		page := &doc.Layout.Page[pi]
		ps := &page.PrintSpace

		// Group blocks by "category" (label)
		var srcBlocks []*TextBlock
		var tgtBlocks []*TextBlock

		for i := range ps.TextBlocks {
			tb := &ps.TextBlocks[i]
			// TAGREFS may be space-separated
			tagIDs := strings.Fields(tb.TagRefs)
			for _, id := range tagIDs {
				label := idToLabel[id]
				if label == "" {
					continue
				}
				if label == stc.StretchCategory {
					srcBlocks = append(srcBlocks, tb)
				} else if label == stc.Towards {
					tgtBlocks = append(tgtBlocks, tb)
				}
			}
		}

		if len(srcBlocks) == 0 || len(tgtBlocks) == 0 {
			continue
		}

		for _, src := range srcBlocks {
			if src == nil || src.Width <= 0 || src.Height <= 0 {
				continue
			}

			// Which sides should be processed
			var stretchingSides []annotationrules.Side
			switch stc.ContactSide {
			case annotationrules.SideAll:
				stretchingSides = []annotationrules.Side{annotationrules.SideLeft, annotationrules.SideRight, annotationrules.SideTop, annotationrules.SideBottom}
			case annotationrules.SideHorizontal:
				stretchingSides = []annotationrules.Side{annotationrules.SideLeft, annotationrules.SideRight}
			case annotationrules.SideVertical:
				stretchingSides = []annotationrules.Side{annotationrules.SideTop, annotationrules.SideBottom}
			default:
				stretchingSides = []annotationrules.Side{stc.ContactSide}
			}

			for _, side := range stretchingSides {
				tgt := chooseTargetBlock(src, tgtBlocks, side, stc.ContactType)
				if tgt == nil {
					continue
				}
				if tgt.Width <= 0 || tgt.Height <= 0 {
					continue
				}

				// Use float64 for geometry then round back to int.
				sx := float64(src.HPOS)
				sy := float64(src.VPOS)
				sw := float64(src.Width)
				sh := float64(src.Height)

				tx := float64(tgt.HPOS)
				ty := float64(tgt.VPOS)
				tw := float64(tgt.Width)
				th := float64(tgt.Height)

				if sw <= 0 || sh <= 0 || tw <= 0 || th <= 0 {
					continue
				}

				newX, newY := sx, sy
				newW, newH := sw, sh

				switch stc.ContactType {
				case annotationrules.ContactTypeInner:
					switch side {
					case annotationrules.SideLeft:
						right := sx + sw
						newX = tx
						newW = right - newX
					case annotationrules.SideRight:
						right := tx + tw
						newW = right - sx
					case annotationrules.SideTop:
						bottom := sy + sh
						newY = ty
						newH = bottom - newY
					case annotationrules.SideBottom:
						bottom := ty + th
						newH = bottom - sy
					default:
						return fmt.Errorf("unsupported contact side %q on page %s", side, page.ID)
					}
				case annotationrules.ContactTypeOuter:
					switch side {
					case annotationrules.SideLeft:
						newW = tx - sx
					case annotationrules.SideRight:
						right := sx + sw
						left := tx + tw
						newW = right - left
						newX = left
					case annotationrules.SideTop:
						newH = ty - sy
					case annotationrules.SideBottom:
						bottom := sy + sh
						top := ty + th
						newH = bottom - top
						newY = top
					default:
						return fmt.Errorf("unsupported contact side %q on page %s", side, page.ID)
					}
				default:
					return fmt.Errorf("unsupported contact type %q", stc.ContactType)
				}

				// Guard against invalid geometry.
				if newW <= 0 || newH <= 0 {
					continue
				}

				// Commit back to ints
				src.HPOS = math.Round(newX)
				src.VPOS = math.Round(newY)
				src.Width = math.Round(newW)
				src.Height = math.Round(newH)

				// Also update polygon to remain a rectangle that matches HPOS/VPOS/WIDTH/HEIGHT.
				updateTextBlockPolygon(src)
			}
		}
	}

	return nil
}

// chooseTargetBlock mirrors chooseTarget but operates on ALTO TextBlocks.
func chooseTargetBlock(src *TextBlock, targets []*TextBlock, side annotationrules.Side, contactType annotationrules.ContactType) *TextBlock {
	if src == nil || len(targets) == 0 {
		return nil
	}
	if src.Width <= 0 || src.Height <= 0 {
		return nil
	}

	sx := float64(src.HPOS)
	sy := float64(src.VPOS)
	sw := float64(src.Width)
	sh := float64(src.Height)

	if sw <= 0 || sh <= 0 {
		return nil
	}

	sLeft, sTop := sx, sy
	sRight, sBottom := sx+sw, sy+sh

	var best *TextBlock
	var bestDist float64
	hasBest := false

	for _, tgt := range targets {
		if tgt == nil || tgt.Width <= 0 || tgt.Height <= 0 {
			continue
		}

		tx := float64(tgt.HPOS)
		ty := float64(tgt.VPOS)
		tw := float64(tgt.Width)
		th := float64(tgt.Height)

		if tw <= 0 || th <= 0 {
			continue
		}

		tLeft, tTop := tx, ty
		tRight, tBottom := tx+tw, ty+th

		match := false
		dist := 0.0

		switch contactType {
		case annotationrules.ContactTypeInner:
			switch side {
			case annotationrules.SideTop:
				if intervalsIntersect(sLeft, sRight, tLeft, tRight) &&
					tTop < sTop && sTop < tBottom {
					match = true
					dist = sTop - tTop
					if dist < 0 {
						match = false
					}
				}
			case annotationrules.SideBottom:
				if intervalsIntersect(sLeft, sRight, tLeft, tRight) &&
					tTop < sBottom && sBottom < tBottom {
					match = true
					dist = tBottom - sBottom
					if dist < 0 {
						match = false
					}
				}
			case annotationrules.SideLeft:
				if intervalsIntersect(sTop, sBottom, tTop, tBottom) &&
					tLeft < sLeft && sLeft < tRight {
					match = true
					dist = sLeft - tLeft
					if dist < 0 {
						match = false
					}
				}
			case annotationrules.SideRight:
				if intervalsIntersect(sTop, sBottom, tTop, tBottom) &&
					tLeft < sRight && sRight < tRight {
					match = true
					dist = tRight - sRight
					if dist < 0 {
						match = false
					}
				}
			}

		case annotationrules.ContactTypeOuter:
			switch side {
			case annotationrules.SideTop:
				if intervalsIntersect(sLeft, sRight, tLeft, tRight) &&
					tBottom <= sTop {
					match = true
					dist = sTop - tBottom
					if dist < 0 {
						match = false
					}
				}
			case annotationrules.SideBottom:
				if intervalsIntersect(sLeft, sRight, tLeft, tRight) &&
					tTop >= sBottom {
					match = true
					dist = tTop - sBottom
					if dist < 0 {
						match = false
					}
				}
			case annotationrules.SideLeft:
				if intervalsIntersect(sTop, sBottom, tTop, tBottom) &&
					tRight <= sLeft {
					match = true
					dist = sLeft - tRight
					if dist < 0 {
						match = false
					}
				}
			case annotationrules.SideRight:
				if intervalsIntersect(sTop, sBottom, tTop, tBottom) &&
					tLeft >= sRight {
					match = true
					dist = tLeft - sRight
					if dist < 0 {
						match = false
					}
				}
			}
		}

		if !match {
			continue
		}

		if !hasBest {
			hasBest = true
			bestDist = dist
			best = tgt
			continue
		}

		switch contactType {
		case annotationrules.ContactTypeInner:
			if dist > bestDist {
				bestDist = dist
				best = tgt
			}
		case annotationrules.ContactTypeOuter:
			if dist < bestDist {
				bestDist = dist
				best = tgt
			}
		}
	}

	return best
}

// intervalsIntersect is the same helper as in the COCO version.
func intervalsIntersect(a1, a2, b1, b2 float64) bool {
	return a1 < b2 && b1 < a2
}

// updateTextBlockPolygon rewrites the Polygon POINTS to match HPOS/VPOS/WIDTH/HEIGHT
// as a simple axis-aligned rectangle.
func updateTextBlockPolygon(tb *TextBlock) {
	if tb == nil {
		return
	}
	x := tb.HPOS
	y := tb.VPOS
	w := tb.Width
	h := tb.Height

	right := x + w
	bottom := y + h

	tb.Shape.Polygon.Points = fmt.Sprintf("%.0f %.0f %.0f %.0f %.0f %.0f %.0f %.0f %.0f %.0f",
		x, y,
		right, y,
		right, bottom,
		x, bottom,
		x, y,
	)
}
