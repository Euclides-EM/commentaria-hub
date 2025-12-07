package alto

import (
	"fmt"
	"strings"
)

// FixNoOverlap ensures that for the given category labels, regions from the same category
// do not overlap. For each overlapping group of blocks in a category it either:
//   - replaces them with a single big bounding rectangle, or
//   - replaces them with several disjoint rectangles covering the exact union,
//
// depending on the "precision" parameter.
//
// precision is an area tolerance (in pixels^2). If
//
//	boundingBoxArea - unionArea <= precision
//
// then the group is replaced by one big block, otherwise by several disjoint blocks.
func FixNoOverlap(a *Alto, categories []string, precision int) error {
	if a == nil || len(categories) == 0 {
		return nil
	}
	if precision < 0 {
		precision = 0
	}

	// Build fast lookup for requested category LABELs
	catSet := make(map[string]struct{}, len(categories))
	for _, c := range categories {
		catSet[c] = struct{}{}
	}

	// Map LABEL -> set of IDs (TAGREFS) that correspond to that label
	labelToIDs := make(map[string]map[string]struct{})
	for _, ot := range a.Tags.OtherTags {
		if _, ok := catSet[ot.Label]; !ok {
			continue
		}
		m, ok := labelToIDs[ot.Label]
		if !ok {
			m = make(map[string]struct{})
			labelToIDs[ot.Label] = m
		}
		m[ot.ID] = struct{}{}
	}

	if len(labelToIDs) == 0 {
		return nil
	}

	for pi := range a.Layout.Page {
		page := &a.Layout.Page[pi]
		blocks := page.PrintSpace.TextBlocks
		if len(blocks) == 0 {
			continue
		}

		n := len(blocks)
		handled := make([]bool, n)
		var newBlocks []TextBlock

		for _, label := range categories {
			idSet, ok := labelToIDs[label]
			if !ok || len(idSet) == 0 {
				continue
			}

			// Collect indices of blocks on this page that belong to this label
			var idxs []int
			for i, b := range blocks {
				if handled[i] {
					continue
				}
				if hasAnyTagRef(b.TagRefs, idSet) {
					idxs = append(idxs, i)
				}
			}
			if len(idxs) <= 1 {
				continue
			}

			// Build rects for these blocks

			rects := make([]rect, len(idxs))
			for i, bi := range idxs {
				b := blocks[bi]
				rects[i] = rect{
					blockIdx: bi,
					x:        b.HPOS,
					y:        b.VPOS,
					w:        b.Width,
					h:        b.Height,
				}
			}

			// Find connected components of overlapping rects
			visited := make([]bool, len(rects))

			for i := range rects {
				if visited[i] {
					continue
				}
				// BFS/DFS from i
				queue := []int{i}
				visited[i] = true
				var group []int

				for len(queue) > 0 {
					r := queue[0]
					queue = queue[1:]
					group = append(group, r)

					for j := range rects {
						if visited[j] {
							continue
						}
						if rectsOverlapNoTol(rects[r], rects[j]) {
							visited[j] = true
							queue = append(queue, j)
						}
					}
				}

				if len(group) <= 1 {
					continue // no overlapping cluster
				}

				// Compute disjoint union rectangles and areas
				var groupRects []rect
				for _, g := range group {
					groupRects = append(groupRects, rects[g])
				}
				unionRects, unionArea := decomposeRectUnion(groupRects)

				// Compute bounding box of the entire group
				minX, minY, maxX, maxY := bboxOfRects(groupRects)
				bboxArea := (maxX - minX) * (maxY - minY)

				// Representative block to clone base fields from
				repBlockIdx := groupRects[0].blockIdx
				rep := blocks[repBlockIdx] // copy by value

				// Mark all originals as handled (to be removed)
				for _, g := range group {
					handled[rects[g].blockIdx] = true
				}

				if bboxArea-unionArea <= precision {
					// Single big block is good enough
					rep.HPOS = minX
					rep.VPOS = minY
					rep.Width = maxX - minX
					rep.Height = maxY - minY
					rep.Shape.Polygon.Points = rectToPolygonPoints(minX, minY, maxX, maxY)
					newBlocks = append(newBlocks, rep)
				} else {
					// Use several disjoint rectangles
					for idx, r := range unionRects {
						nb := rep
						nb.HPOS = r.x
						nb.VPOS = r.y
						nb.Width = r.w
						nb.Height = r.h
						nb.Shape.Polygon.Points = rectToPolygonPoints(r.x, r.y, r.x+r.w, r.y+r.h)
						if idx > 0 {
							nb.ID = fmt.Sprintf("%s_%d", rep.ID, idx)
						}
						newBlocks = append(newBlocks, nb)
					}
				}
			}
		}

		// Rebuild block list: keep all unhandled, then add newly created blocks
		if len(newBlocks) > 0 {
			var finalBlocks []TextBlock
			finalBlocks = make([]TextBlock, 0, len(blocks)-countTrue(handled)+len(newBlocks))
			for i, b := range blocks {
				if !handled[i] {
					finalBlocks = append(finalBlocks, b)
				}
			}
			finalBlocks = append(finalBlocks, newBlocks...)
			page.PrintSpace.TextBlocks = finalBlocks
		}
	}
	return nil
}

// hasAnyTagRef returns true if TAGREFS string contains any ID from ids.
func hasAnyTagRef(tagRefs string, ids map[string]struct{}) bool {
	if tagRefs == "" {
		return false
	}
	for _, p := range strings.Fields(tagRefs) {
		if _, ok := ids[p]; ok {
			return true
		}
	}
	return false
}

func rectsOverlapNoTol(a, b rect) bool {
	ax1 := a.x
	ay1 := a.y
	ax2 := a.x + a.w
	ay2 := a.y + a.h

	bx1 := b.x
	by1 := b.y
	bx2 := b.x + b.w
	by2 := b.y + b.h

	return ax1 < bx2 && ax2 > bx1 && ay1 < by2 && ay2 > by1
}

// decomposeRectUnion decomposes the union of axis-aligned rectangles into a set of
// disjoint rectangles that cover exactly the union. It returns those rectangles
// and the total area of the union.
func decomposeRectUnion(in []rect) (out []rect, area int) {
	if len(in) == 0 {
		return nil, 0
	}

	// Collect unique vertical edges
	var xs []int
	xSeen := make(map[int]struct{})
	for _, r := range in {
		x1 := r.x
		x2 := r.x + r.w
		if _, ok := xSeen[x1]; !ok {
			xSeen[x1] = struct{}{}
			xs = append(xs, x1)
		}
		if _, ok := xSeen[x2]; !ok {
			xSeen[x2] = struct{}{}
			xs = append(xs, x2)
		}
	}
	// sort xs
	for i := 0; i < len(xs); i++ {
		for j := i + 1; j < len(xs); j++ {
			if xs[j] < xs[i] {
				xs[i], xs[j] = xs[j], xs[i]
			}
		}
	}

	var res []rect
	totalArea := 0

	// Sweep vertical strips between consecutive x-coordinates
	for i := 0; i < len(xs)-1; i++ {
		x1 := xs[i]
		x2 := xs[i+1]
		if x2 <= x1 {
			continue
		}

		// Collect y-intervals of rects that intersect this strip
		type interval struct {
			y1, y2 int
		}
		var intervals []interval
		for _, r := range in {
			rx1 := r.x
			rx2 := r.x + r.w
			if rx1 >= x2 || rx2 <= x1 {
				continue
			}
			y1 := r.y
			y2 := r.y + r.h
			intervals = append(intervals, interval{y1: y1, y2: y2})
		}
		if len(intervals) == 0 {
			continue
		}

		// Merge y-intervals
		for i := 0; i < len(intervals); i++ {
			for j := i + 1; j < len(intervals); j++ {
				if intervals[j].y1 < intervals[i].y1 {
					intervals[i], intervals[j] = intervals[j], intervals[i]
				}
			}
		}

		merged := make([]interval, 0, len(intervals))
		cur := intervals[0]
		for _, iv := range intervals[1:] {
			if iv.y1 <= cur.y2 {
				if iv.y2 > cur.y2 {
					cur.y2 = iv.y2
				}
			} else {
				merged = append(merged, cur)
				cur = iv
			}
		}
		merged = append(merged, cur)

		// Produce rectangles for this strip
		for _, iv := range merged {
			w := x2 - x1
			h := iv.y2 - iv.y1
			if w <= 0 || h <= 0 {
				continue
			}
			res = append(res, rect{
				blockIdx: -1,
				x:        x1,
				y:        iv.y1,
				w:        w,
				h:        h,
			})
			totalArea += w * h
		}
	}

	out = make([]rect, len(res))
	for i, r := range res {
		out[i] = rect{
			blockIdx: r.blockIdx,
			x:        r.x,
			y:        r.y,
			w:        r.w,
			h:        r.h,
		}
	}

	return out, totalArea
}

func bboxOfRects(rs []rect) (minX, minY, maxX, maxY int) {
	if len(rs) == 0 {
		return 0, 0, 0, 0
	}
	minX = rs[0].x
	minY = rs[0].y
	maxX = rs[0].x + rs[0].w
	maxY = rs[0].y + rs[0].h
	for _, r := range rs[1:] {
		if r.x < minX {
			minX = r.x
		}
		if r.y < minY {
			minY = r.y
		}
		if r.x+r.w > maxX {
			maxX = r.x + r.w
		}
		if r.y+r.h > maxY {
			maxY = r.y + r.h
		}
	}
	return
}

func rectToPolygonPoints(minX, minY, maxX, maxY int) string {
	return fmt.Sprintf(
		"%d %d %d %d %d %d %d %d %d %d",
		minX, minY,
		maxX, minY,
		maxX, maxY,
		minX, maxY,
		minX, minY,
	)
}

func countTrue(bs []bool) int {
	n := 0
	for _, v := range bs {
		if v {
			n++
		}
	}
	return n
}

type rect struct {
	blockIdx int
	x, y     int
	w, h     int
}
