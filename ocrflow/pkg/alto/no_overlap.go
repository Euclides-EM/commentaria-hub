package alto

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/geo"
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
func FixNoOverlap(a *Alto, categories []string, precision float64) error {
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
					blockIdx:  bi,
					Rectangle: *geo.RectangleFromLeftBottomCorner(b.HPOS, b.VPOS, b.Width, b.Height),
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
						nb.HPOS = r.MinX
						nb.VPOS = r.MinY
						nb.Width = r.MaxX - r.MinX
						nb.Height = r.MaxY - r.MinY
						nb.Shape.Polygon.Points = rectToPolygonPoints(r.MinX, r.MinY, r.MaxX, r.MaxY)
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
	ax1 := a.MinX
	ay1 := a.MinY
	ax2 := a.MaxX
	ay2 := a.MaxY

	bx1 := b.MinX
	by1 := b.MinY
	bx2 := b.MaxX
	by2 := b.MaxY

	return ax1 < bx2 && ax2 > bx1 && ay1 < by2 && ay2 > by1
}

// decomposeRectUnion decomposes the union of axis-aligned rectangles into a set of
// disjoint rectangles that cover exactly the union. It returns those rectangles
// and the total area of the union.
func decomposeRectUnion(in []rect) (out []rect, area float64) {
	if len(in) == 0 {
		return nil, 0
	}

	// Collect unique vertical edges
	var xs []float64
	xSeen := make(map[float64]struct{})
	for _, r := range in {
		x1 := r.MinX
		x2 := r.MaxX
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
	totalArea := 0.0

	// Sweep vertical strips between consecutive x-coordinates
	for i := 0; i < len(xs)-1; i++ {
		x1 := xs[i]
		x2 := xs[i+1]
		if x2 <= x1 {
			continue
		}

		// Collect y-intervals of rects that intersect this strip
		type interval struct {
			y1, y2 float64
		}
		var intervals []interval
		for _, r := range in {
			rx1 := r.MinX
			rx2 := r.MaxX
			if rx1 >= x2 || rx2 <= x1 {
				continue
			}
			y1 := r.MinY
			y2 := r.MaxY
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
			r := rect{
				blockIdx: -1,
				Rectangle: geo.Rectangle{
					MinX: x1,
					MinY: iv.y1,
					MaxX: x2,
					MaxY: iv.y2,
				},
			}
			res = append(res, r)
			totalArea += r.Area()
		}
	}

	out = make([]rect, len(res))
	for i, r := range res {
		out[i] = rect{
			blockIdx: r.blockIdx,
			Rectangle: geo.Rectangle{
				MinX: r.MinX,
				MinY: r.MinY,
				MaxX: r.MaxX,
				MaxY: r.MaxY,
			},
		}
	}

	return out, totalArea
}

func bboxOfRects(rs []rect) (minX, minY, maxX, maxY float64) {
	if len(rs) == 0 {
		return 0, 0, 0, 0
	}
	minX = rs[0].MinX
	minY = rs[0].MinY
	maxX = rs[0].MaxX
	maxY = rs[0].MaxY
	for _, r := range rs[1:] {
		if r.MinX < minX {
			minX = r.MinX
		}
		if r.MinY < minY {
			minY = r.MinY
		}
		if r.MaxX > maxX {
			maxX = r.MaxX
		}
		if r.MaxY > maxY {
			maxY = r.MaxY
		}
	}
	return
}

func rectToPolygonPoints(minX, minY, maxX, maxY float64) string {
	return fmt.Sprintf(
		"%d %d %d %d %d %d %d %d %d %d",
		int(minX), int(minY),
		int(maxX), int(minY),
		int(maxX), int(maxY),
		int(minX), int(maxY),
		int(minX), int(minY),
	)
}

func polygonPointsToRect(polygonPoints string) (*geo.Rectangle, error) {
	fields := strings.Fields(polygonPoints)
	if len(fields) != 10 {
		return nil, fmt.Errorf("expected 10 coordinates in polygon points, got %d", len(fields))
	}

	minX := float64(0)
	minY := float64(0)
	maxX := float64(0)
	maxY := float64(0)

	for i := 0; i < len(fields); i += 2 {
		x, err1 := strconv.ParseFloat(fields[i], 64)
		y, err2 := strconv.ParseFloat(fields[i+1], 64)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("invalid coordinate in polygon points: %v, %v", err1, err2)
		}
		if i == 0 {
			minX = x
			maxX = x
			minY = y
			maxY = y
		} else {
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	return &geo.Rectangle{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}, nil
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
	geo.Rectangle
}
