package alto

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/geo"
	"github.com/samber/lo"
)

// ReassignTextLinesByTolerance moves TextLine elements from TextBlocks tagged fromCat
// into TextBlocks tagged toCat when they fit within tolerance.
// fromCat/toCat can be either Tag ID (e.g. "BT224") or Tag LABEL (e.g. "MainZone").
// precisionPx expands candidate toCat blocks by this many pixels in all directions.
// minOverlap is a ratio in (0..1], e.g. 0.80 means 80% of the line bbox must overlap.
func ReassignTextLinesByTolerance(a *Alto, fromCat, toCat string, precisionPx float64, minOverlap float64) (moved int, err error) {
	if minOverlap <= 0 || minOverlap > 1 {
		return 0, fmt.Errorf("minOverlap must be in (0, 1], got %v", minOverlap)
	}

	// Resolve label -> ID if needed (and allow already-provided IDs).
	fromID, err := resolveTagID(a, fromCat)
	if err != nil {
		return 0, fmt.Errorf("fromCat: %w", err)
	}
	toID, err := resolveTagID(a, toCat)
	if err != nil {
		return 0, fmt.Errorf("toCat: %w", err)
	}

	type blockRef struct {
		pageIdx int
		tbIdx   int
		rect    *geo.Rectangle
	}

	var targets []blockRef
	var sources []blockRef

	// Collect sources + targets as references into a.Layout.
	for pi := range a.Layout.Page {
		tbs := a.Layout.Page[pi].PrintSpace.TextBlocks
		for tbi := range tbs {
			tb := a.Layout.Page[pi].PrintSpace.TextBlocks[tbi]
			tokens := strings.Fields(strings.TrimSpace(tb.TagRefs))

			if lo.Contains(tokens, fromID) {
				sources = append(sources, blockRef{pageIdx: pi, tbIdx: tbi})
			}

			if !lo.Contains(tokens, toID) {
				continue
			}
			r, ok := rectForElement(tb)
			if !ok {
				continue
			}
			targets = append(targets, blockRef{pageIdx: pi, tbIdx: tbi, rect: r})
		}
	}

	if len(targets) == 0 {
		return 0, nil
	}

	// For each source block, scan its lines and move lines that fit best into a target.
	for _, src := range sources {
		srcTB := &a.Layout.Page[src.pageIdx].PrintSpace.TextBlocks[src.tbIdx]

		// Iterate over a snapshot of IDs so deletions do not break iteration.
		srcLines := append([]TextLine(nil), srcTB.Lines...)

		for _, tl := range srcLines {
			lineRect := geo.RectangleFromLeftBottomCorner(tl.HPOS, tl.VPOS, tl.Width, tl.Height)

			bestIdx := -1
			bestScore := 0.0

			for i, t := range targets {
				expanded := t.rect.Expand(precisionPx)

				if expanded.Contains(lineRect) {
					bestIdx = i
					bestScore = 1.0
					break
				}

				score := lineRect.OverlapRatio(expanded)
				if score > bestScore {
					bestScore = score
					bestIdx = i
				}
			}

			if bestIdx < 0 || bestScore < minOverlap {
				continue
			}

			// Remove from source (in place)
			srcTB.Lines = lo.Filter(srcTB.Lines, func(l TextLine, _ int) bool {
				return l.ID != tl.ID
			})

			// Append to target (in place)
			tref := targets[bestIdx]
			dstTB := &a.Layout.Page[tref.pageIdx].PrintSpace.TextBlocks[tref.tbIdx]
			dstTB.Lines = append(dstTB.Lines, tl)

			moved++
		}
	}

	// Keep reading order inside each target block: sort by (VPOS, HPOS)
	for _, t := range targets {
		tb := &a.Layout.Page[t.pageIdx].PrintSpace.TextBlocks[t.tbIdx]
		sortTextLinesByPosInPlace(tb)
	}

	return moved, nil
}

// resolveTagID accepts either an ID (e.g. "BT224") or a LABEL (e.g. "MainZone") and returns the ID.
func resolveTagID(a *Alto, s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty tag")
	}

	// Fast path: looks like an ID and exists.
	if strings.HasPrefix(s, "BT") {
		for _, t := range a.Tags.OtherTags {
			if strings.TrimSpace(t.ID) == s {
				return t.ID, nil
			}
		}
		// If it starts with BT but isn't present, still error (safer than silently accepting).
		return "", fmt.Errorf("tag id %q not found in <Tags>", s)
	}

	// Otherwise treat as LABEL.
	for _, t := range a.Tags.OtherTags {
		if strings.TrimSpace(t.Label) == s {
			return t.ID, nil
		}
	}

	return "", fmt.Errorf("tag label %q not found in <Tags>", s)
}

func sortTextLinesByPosInPlace(tb *TextBlock) {
	if len(tb.Lines) <= 1 {
		return
	}
	sort.SliceStable(tb.Lines, func(i, j int) bool {
		if tb.Lines[i].VPOS == tb.Lines[j].VPOS {
			return tb.Lines[i].HPOS < tb.Lines[j].HPOS
		}
		return tb.Lines[i].VPOS < tb.Lines[j].VPOS
	})
}

func rectForElement(el TextBlock) (*geo.Rectangle, bool) {
	// Prefer HPOS/VPOS/WIDTH/HEIGHT.
	if el.Width >= 0 && el.Height >= 0 {
		return geo.RectangleFromLeftBottomCorner(el.HPOS, el.VPOS, el.Width, el.Height), true
	}

	// Fallback: Shape/Polygon POINTS="x y x y ..."
	r, err := polygonPointsToRect(el.Shape.Polygon.Points)
	if err != nil {
		log.Printf("polygonPointsToRect failed: %v", err)
		return nil, false
	}
	return r, true
}
