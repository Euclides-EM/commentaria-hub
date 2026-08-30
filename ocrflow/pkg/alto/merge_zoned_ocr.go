package alto

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// MergeStats summarizes one merge operation.
type MergeStats struct {
	Pages           int
	InputLines      int
	OutputLines     int
	RemovedLines    int
	SplitLines      int
	ReassignedLines int
	RemovedRunes    int
	OutputRunes     int
}

// LineReassignment moves merged lines from one category into an overlapping
// target category. Reassignments are applied in order, allowing earlier rules
// to take precedence over later rules with the same source category.
type LineReassignment struct {
	FromCategory string  `json:"from_category"`
	ToCategory   string  `json:"to_category"`
	PrecisionPx  float64 `json:"precision_px"`
	MinOverlap   float64 `json:"min_overlap"`
}

// MergeZonedOCRDirs merges a flat segmentation ALTO directory with an OCR ALTO
// directory laid out as page-NNNN/original.xml. It also accepts flat OCR input.
// Output files use the segmentation page names in page-NNNN/original.xml
// directories and the segmentation coordinate system.
func MergeZonedOCRDirs(segmentationDir, ocrDir, outputDir string, includeCategories, ignoreCategories []string) (MergeStats, error) {
	return MergeZonedOCRDirsWithReassignments(segmentationDir, ocrDir, outputDir, includeCategories, ignoreCategories, nil)
}

// MergeZonedOCRDirsWithReassignments behaves like MergeZonedOCRDirs and then
// applies the supplied ordered category reassignment rules to every page.
func MergeZonedOCRDirsWithReassignments(segmentationDir, ocrDir, outputDir string, includeCategories, ignoreCategories []string, reassignments []LineReassignment) (MergeStats, error) {
	var stats MergeStats
	if len(includeCategories) == 0 {
		return stats, fmt.Errorf("at least one include category is required")
	}
	for i, rule := range reassignments {
		if err := validateLineReassignment(rule); err != nil {
			return stats, fmt.Errorf("reassignment %d: %w", i+1, err)
		}
	}
	for _, pair := range [][2]string{{segmentationDir, "segmentation"}, {ocrDir, "OCR"}} {
		info, err := os.Stat(pair[0])
		if err != nil {
			return stats, fmt.Errorf("stat %s directory %q: %w", pair[1], pair[0], err)
		}
		if !info.IsDir() {
			return stats, fmt.Errorf("%s path %q is not a directory", pair[1], pair[0])
		}
	}
	for _, input := range []string{segmentationDir, ocrDir} {
		same, err := samePath(input, outputDir)
		if err != nil {
			return stats, err
		}
		if same {
			return stats, fmt.Errorf("refusing to overwrite input directory %q", input)
		}
	}

	segPages, err := discoverALTOPages(segmentationDir, true)
	if err != nil {
		return stats, fmt.Errorf("discover segmentation ALTO: %w", err)
	}
	ocrPages, err := discoverALTOPages(ocrDir, false)
	if err != nil {
		return stats, fmt.Errorf("discover OCR ALTO: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return stats, fmt.Errorf("create output directory %q: %w", outputDir, err)
	}

	keys := make([]string, 0, len(segPages))
	for key := range segPages {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		ocrPath, ok := ocrPages[key]
		if !ok {
			return stats, fmt.Errorf("no OCR ALTO page matching %q", filepath.Base(segPages[key]))
		}
		seg, err := LoadFromFile(segPages[key])
		if err != nil {
			return stats, fmt.Errorf("load segmentation page %q: %w", segPages[key], err)
		}
		ocr, err := LoadFromFile(ocrPath)
		if err != nil {
			return stats, fmt.Errorf("load OCR page %q: %w", ocrPath, err)
		}
		pageStats, err := MergeZonedOCR(seg, ocr, includeCategories, ignoreCategories)
		if err != nil {
			return stats, fmt.Errorf("merge page %q: %w", key, err)
		}
		for _, rule := range reassignments {
			if !hasCategory(seg, rule.FromCategory) || !hasCategory(seg, rule.ToCategory) {
				continue
			}
			moved, err := ReassignTextLinesByTolerance(seg, rule.FromCategory, rule.ToCategory, rule.PrecisionPx, rule.MinOverlap)
			if err != nil {
				return stats, fmt.Errorf("reassign page %q from %q to %q: %w", key, rule.FromCategory, rule.ToCategory, err)
			}
			pageStats.ReassignedLines += moved
		}
		stats.add(pageStats)
		stats.Pages++
		pageDir := strings.TrimSuffix(filepath.Base(segPages[key]), filepath.Ext(segPages[key]))
		outPath := filepath.Join(outputDir, pageDir, "original.xml")
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return stats, fmt.Errorf("create output page directory %q: %w", filepath.Dir(outPath), err)
		}
		if err := SaveToFile(seg, outPath); err != nil {
			return stats, fmt.Errorf("save merged page %q: %w", outPath, err)
		}
	}
	return stats, nil
}

func (s *MergeStats) add(other MergeStats) {
	s.InputLines += other.InputLines
	s.OutputLines += other.OutputLines
	s.RemovedLines += other.RemovedLines
	s.SplitLines += other.SplitLines
	s.ReassignedLines += other.ReassignedLines
	s.RemovedRunes += other.RemovedRunes
	s.OutputRunes += other.OutputRunes
}

func validateLineReassignment(rule LineReassignment) error {
	if strings.TrimSpace(rule.FromCategory) == "" {
		return fmt.Errorf("from_category is required")
	}
	if strings.TrimSpace(rule.ToCategory) == "" {
		return fmt.Errorf("to_category is required")
	}
	if rule.PrecisionPx < 0 {
		return fmt.Errorf("precision_px must be non-negative, got %v", rule.PrecisionPx)
	}
	if rule.MinOverlap <= 0 || rule.MinOverlap > 1 {
		return fmt.Errorf("min_overlap must be in (0, 1], got %v", rule.MinOverlap)
	}
	return nil
}

func hasCategory(doc *Alto, category string) bool {
	category = strings.TrimSpace(category)
	for _, tag := range doc.Tags.OtherTags {
		if strings.TrimSpace(tag.ID) == category || strings.TrimSpace(tag.Label) == category {
			return true
		}
	}
	return false
}

// MergeZonedOCR modifies segmentation in place. Its zones are retained, all
// existing lines are removed, and scaled OCR lines are clipped into included
// zones with ignored zones acting as holes.
func MergeZonedOCR(segmentation, ocr *Alto, includeCategories, ignoreCategories []string) (MergeStats, error) {
	var stats MergeStats
	if len(segmentation.Layout.Page) != len(ocr.Layout.Page) {
		return stats, fmt.Errorf("page count differs: segmentation=%d OCR=%d", len(segmentation.Layout.Page), len(ocr.Layout.Page))
	}
	include := stringSet(includeCategories)
	ignore := stringSet(ignoreCategories)
	if len(include) == 0 {
		return stats, fmt.Errorf("at least one non-empty include category is required")
	}

	labels := make(map[string]string, len(segmentation.Tags.OtherTags))
	for _, tag := range segmentation.Tags.OtherTags {
		labels[strings.TrimSpace(tag.ID)] = strings.TrimSpace(tag.Label)
	}
	for pi := range segmentation.Layout.Page {
		segPage := &segmentation.Layout.Page[pi]
		ocrPage := &ocr.Layout.Page[pi]
		if segPage.Width <= 0 || segPage.Height <= 0 || ocrPage.Width <= 0 || ocrPage.Height <= 0 {
			return stats, fmt.Errorf("page %d has invalid dimensions: segmentation=%dx%d OCR=%dx%d", pi+1, segPage.Width, segPage.Height, ocrPage.Width, ocrPage.Height)
		}

		var included, ignored []zoneRef
		for bi := range segPage.PrintSpace.TextBlocks {
			block := &segPage.PrintSpace.TextBlocks[bi]
			block.Lines = nil
			categoryLabels := tagLabels(block.TagRefs, labels)
			isIgnored := intersectsSet(categoryLabels, ignore)
			isIncluded := intersectsSet(categoryLabels, include)
			if !isIgnored && !isIncluded {
				continue
			}
			poly, err := elementPolygon(block.Shape.Polygon.Points, block.HPOS, block.VPOS, block.Width, block.Height)
			if err != nil {
				return stats, fmt.Errorf("zone %q: %w", block.ID, err)
			}
			if isIgnored {
				ignored = append(ignored, zoneRef{blockIndex: bi, polygon: poly, area: math.Abs(polygonArea(poly))})
			}
			if isIncluded {
				included = append(included, zoneRef{blockIndex: bi, polygon: poly, area: math.Abs(polygonArea(poly))})
			}
		}
		if len(included) == 0 {
			continue
		}

		sx := float64(segPage.Width) / float64(ocrPage.Width)
		sy := float64(segPage.Height) / float64(ocrPage.Height)
		for _, sourceBlock := range ocrPage.PrintSpace.TextBlocks {
			for _, original := range sourceBlock.Lines {
				stats.InputLines++
				line := scaleTextLine(original, sx, sy)
				fragments, err := clipLineToZones(line, included, ignored)
				if err != nil {
					return stats, fmt.Errorf("line %q: %w", original.ID, err)
				}
				inputRunes := lineRuneCount(line)
				if len(fragments) == 0 {
					stats.RemovedLines++
					stats.RemovedRunes += inputRunes
					continue
				}
				if len(fragments) > 1 || fragments[0].start > geometryEpsilon || fragments[0].end < 1-geometryEpsilon {
					stats.SplitLines++
				}
				outputRunes := 0
				for i, fragment := range fragments {
					clipped, ok := makeLineFragment(line, fragment, i, len(fragments), included[fragment.owner].polygon)
					if !ok {
						continue
					}
					segPage.PrintSpace.TextBlocks[included[fragment.owner].blockIndex].Lines = append(
						segPage.PrintSpace.TextBlocks[included[fragment.owner].blockIndex].Lines, clipped,
					)
					stats.OutputLines++
					outputRunes += lineRuneCount(clipped)
				}
				stats.OutputRunes += outputRunes
				if inputRunes > outputRunes {
					stats.RemovedRunes += inputRunes - outputRunes
				}
			}
		}
		for bi := range segPage.PrintSpace.TextBlocks {
			sortTextLinesByPosInPlace(&segPage.PrintSpace.TextBlocks[bi])
		}
	}
	return stats, nil
}

const geometryEpsilon = 1e-7

type point struct{ x, y float64 }
type zoneRef struct {
	blockIndex int
	polygon    []point
	area       float64
}
type lineFragment struct {
	start, end float64
	owner      int
}

func clipLineToZones(line TextLine, included, ignored []zoneRef) ([]lineFragment, error) {
	baseline, err := lineBaseline(line)
	if err != nil {
		return nil, err
	}
	if len(baseline) < 2 {
		return nil, nil
	}
	lengths, total := cumulativeLengths(baseline)
	if total <= geometryEpsilon {
		return nil, nil
	}
	breaks := []float64{0, 1}
	allZones := append(append([]zoneRef(nil), included...), ignored...)
	for si := 0; si+1 < len(baseline); si++ {
		segLen := lengths[si+1] - lengths[si]
		if segLen <= geometryEpsilon {
			continue
		}
		for _, zone := range allZones {
			for ei := range zone.polygon {
				a := zone.polygon[ei]
				b := zone.polygon[(ei+1)%len(zone.polygon)]
				if u, ok := segmentIntersectionParameter(baseline[si], baseline[si+1], a, b); ok {
					breaks = append(breaks, (lengths[si]+u*segLen)/total)
				}
			}
		}
	}
	breaks = uniqueSorted(breaks)
	var result []lineFragment
	for i := 0; i+1 < len(breaks); i++ {
		a, b := breaks[i], breaks[i+1]
		if b-a <= geometryEpsilon {
			continue
		}
		mid := pointAtDistance(baseline, lengths, total*(a+b)/2)
		blocked := false
		for _, zone := range ignored {
			if pointInPolygon(mid, zone.polygon) {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}
		owner := -1
		for zi, zone := range included {
			if pointInPolygon(mid, zone.polygon) && (owner < 0 || zone.area < included[owner].area) {
				owner = zi
			}
		}
		if owner < 0 {
			continue
		}
		if len(result) > 0 && result[len(result)-1].owner == owner && math.Abs(result[len(result)-1].end-a) <= geometryEpsilon {
			result[len(result)-1].end = b
		} else {
			result = append(result, lineFragment{start: a, end: b, owner: owner})
		}
	}
	return result, nil
}

func makeLineFragment(line TextLine, fragment lineFragment, partIndex, partCount int, ownerPolygon []point) (TextLine, bool) {
	baseline, _ := lineBaseline(line)
	lengths, total := cumulativeLengths(baseline)
	clippedBaseline := clipPolyline(baseline, lengths, total*fragment.start, total*fragment.end)
	if len(clippedBaseline) < 2 {
		return TextLine{}, false
	}

	linePoly, err := elementPolygon(line.Shape.Polygon.Points, line.HPOS, line.VPOS, line.Width, line.Height)
	if err != nil {
		return TextLine{}, false
	}
	startPoint := clippedBaseline[0]
	endPoint := clippedBaseline[len(clippedBaseline)-1]
	minX, maxX := math.Min(startPoint.x, endPoint.x), math.Max(startPoint.x, endPoint.x)
	// Baselines are expected to run left-to-right. The vertical slab ensures an
	// ignored-zone crossing creates separate, non-overlapping line polygons.
	linePoly = clipPolygonHalfPlane(linePoly, func(p point) bool { return p.x >= minX-geometryEpsilon }, func(a, b point) point { return intersectVertical(a, b, minX) })
	linePoly = clipPolygonHalfPlane(linePoly, func(p point) bool { return p.x <= maxX+geometryEpsilon }, func(a, b point) point { return intersectVertical(a, b, maxX) })
	if isConvex(ownerPolygon) {
		linePoly = clipPolygonConvex(linePoly, ownerPolygon)
	}
	if len(linePoly) < 3 {
		return TextLine{}, false
	}

	out := line
	if partCount > 1 || fragment.start > geometryEpsilon || fragment.end < 1-geometryEpsilon {
		out.ID = fmt.Sprintf("%s__part_%d", line.ID, partIndex+1)
	}
	out.Baseline = formatPoints(clippedBaseline, false)
	out.Shape.Polygon.Points = formatPoints(linePoly, true)
	setLineBBoxFromPolygon(&out, linePoly)
	out.Strings = clipStrings(line.Strings, fragment.start, fragment.end, line.HPOS, line.Width, out.HPOS, out.Width)
	if lineRuneCount(out) == 0 {
		return TextLine{}, false
	}
	return out, true
}

func clipStrings(stringsIn []String, start, end, lineX, lineWidth, outX, outWidth float64) []String {
	if len(stringsIn) == 0 {
		return nil
	}
	var result []String
	for _, token := range stringsIn {
		runes := []rune(token.Content)
		if len(runes) == 0 {
			continue
		}
		t0, t1 := start, end
		if token.Width > 0 && lineWidth > 0 {
			tokenStart := (token.HPOS - lineX) / lineWidth
			tokenEnd := (token.HPOS + token.Width - lineX) / lineWidth
			visibleStart := math.Max(start, tokenStart)
			visibleEnd := math.Min(end, tokenEnd)
			if visibleEnd <= visibleStart {
				continue
			}
			t0 = (visibleStart - tokenStart) / math.Max(tokenEnd-tokenStart, geometryEpsilon)
			t1 = (visibleEnd - tokenStart) / math.Max(tokenEnd-tokenStart, geometryEpsilon)
		} else {
			t0 = (start - 0) / 1
			t1 = (end - 0) / 1
		}
		from := clampInt(int(math.Round(t0*float64(len(runes)))), 0, len(runes))
		to := clampInt(int(math.Round(t1*float64(len(runes)))), 0, len(runes))
		if to <= from {
			continue
		}
		clipped := token
		clipped.Content = string(runes[from:to])
		if clipped.Width > 0 {
			oldRight := clipped.HPOS + clipped.Width
			clipped.HPOS = roundGeometry(math.Max(clipped.HPOS, outX))
			clipped.Width = roundGeometry(math.Max(0, math.Min(oldRight, outX+outWidth)-clipped.HPOS))
		}
		result = append(result, clipped)
	}
	return result
}

func scaleTextLine(in TextLine, sx, sy float64) TextLine {
	out := in
	out.Strings = append([]String(nil), in.Strings...)
	out.HPOS = roundGeometry(out.HPOS * sx)
	out.VPOS = roundGeometry(out.VPOS * sy)
	out.Width = roundGeometry(out.Width * sx)
	out.Height = roundGeometry(out.Height * sy)
	out.Shape.Polygon.Points = scalePointsString(out.Shape.Polygon.Points, sx, sy)
	out.Baseline = scalePointsString(out.Baseline, sx, sy)
	for i := range out.Strings {
		out.Strings[i].HPOS = roundGeometry(out.Strings[i].HPOS * sx)
		out.Strings[i].VPOS = roundGeometry(out.Strings[i].VPOS * sy)
		out.Strings[i].Width = roundGeometry(out.Strings[i].Width * sx)
		out.Strings[i].Height = roundGeometry(out.Strings[i].Height * sy)
	}
	return out
}

func lineBaseline(line TextLine) ([]point, error) {
	if strings.TrimSpace(line.Baseline) != "" {
		points, err := parsePoints(line.Baseline)
		if err == nil && len(points) >= 2 {
			return points, nil
		}
	}
	y := line.VPOS + line.Height/2
	return []point{{line.HPOS, y}, {line.HPOS + line.Width, y}}, nil
}

func elementPolygon(points string, x, y, width, height float64) ([]point, error) {
	if strings.TrimSpace(points) != "" {
		parsed, err := parsePoints(points)
		if err != nil {
			return nil, err
		}
		if len(parsed) >= 3 {
			if samePoint(parsed[0], parsed[len(parsed)-1]) {
				parsed = parsed[:len(parsed)-1]
			}
			return parsed, nil
		}
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("missing usable polygon and bounding box")
	}
	return []point{{x, y}, {x + width, y}, {x + width, y + height}, {x, y + height}}, nil
}

func parsePoints(raw string) ([]point, error) {
	fields := strings.Fields(raw)
	if len(fields)%2 != 0 {
		return nil, fmt.Errorf("POINTS has an odd number of coordinates")
	}
	result := make([]point, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		x, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid x coordinate %q: %w", fields[i], err)
		}
		y, err := strconv.ParseFloat(fields[i+1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid y coordinate %q: %w", fields[i+1], err)
		}
		result = append(result, point{x, y})
	}
	return result, nil
}

func formatPoints(points []point, closePolygon bool) string {
	if closePolygon && len(points) > 0 && !samePoint(points[0], points[len(points)-1]) {
		points = append(append([]point(nil), points...), points[0])
	}
	parts := make([]string, 0, len(points)*2)
	for _, p := range points {
		parts = append(parts, formatCoordinate(p.x), formatCoordinate(p.y))
	}
	return strings.Join(parts, " ")
}

func formatCoordinate(v float64) string {
	if math.Abs(v-math.Round(v)) < 1e-6 {
		return strconv.FormatInt(int64(math.Round(v)), 10)
	}
	return strconv.FormatFloat(v, 'f', 3, 64)
}

func scalePointsString(raw string, sx, sy float64) string {
	points, err := parsePoints(raw)
	if err != nil {
		return raw
	}
	for i := range points {
		points[i].x *= sx
		points[i].y *= sy
	}
	return formatPoints(points, false)
}

func cumulativeLengths(points []point) ([]float64, float64) {
	lengths := make([]float64, len(points))
	for i := 1; i < len(points); i++ {
		lengths[i] = lengths[i-1] + math.Hypot(points[i].x-points[i-1].x, points[i].y-points[i-1].y)
	}
	if len(lengths) == 0 {
		return lengths, 0
	}
	return lengths, lengths[len(lengths)-1]
}

func pointAtDistance(points []point, lengths []float64, distance float64) point {
	if distance <= 0 {
		return points[0]
	}
	for i := 0; i+1 < len(points); i++ {
		if distance <= lengths[i+1] {
			ratio := (distance - lengths[i]) / math.Max(lengths[i+1]-lengths[i], geometryEpsilon)
			return interpolate(points[i], points[i+1], ratio)
		}
	}
	return points[len(points)-1]
}

func clipPolyline(points []point, lengths []float64, start, end float64) []point {
	result := []point{pointAtDistance(points, lengths, start)}
	for i := 1; i+1 < len(points); i++ {
		if lengths[i] > start+geometryEpsilon && lengths[i] < end-geometryEpsilon {
			result = append(result, points[i])
		}
	}
	result = append(result, pointAtDistance(points, lengths, end))
	return result
}

func segmentIntersectionParameter(p, p2, q, q2 point) (float64, bool) {
	r := point{p2.x - p.x, p2.y - p.y}
	s := point{q2.x - q.x, q2.y - q.y}
	denom := cross(r, s)
	if math.Abs(denom) <= geometryEpsilon {
		return 0, false
	}
	qp := point{q.x - p.x, q.y - p.y}
	t := cross(qp, s) / denom
	u := cross(qp, r) / denom
	return t, t >= -geometryEpsilon && t <= 1+geometryEpsilon && u >= -geometryEpsilon && u <= 1+geometryEpsilon
}

func pointInPolygon(p point, polygon []point) bool {
	inside := false
	for i, j := 0, len(polygon)-1; i < len(polygon); j, i = i, i+1 {
		a, b := polygon[j], polygon[i]
		if pointOnSegment(p, a, b) {
			return true
		}
		if (a.y > p.y) != (b.y > p.y) && p.x < (b.x-a.x)*(p.y-a.y)/(b.y-a.y)+a.x {
			inside = !inside
		}
	}
	return inside
}

func pointOnSegment(p, a, b point) bool {
	return math.Abs(cross(point{p.x - a.x, p.y - a.y}, point{b.x - a.x, b.y - a.y})) < geometryEpsilon &&
		p.x >= math.Min(a.x, b.x)-geometryEpsilon && p.x <= math.Max(a.x, b.x)+geometryEpsilon &&
		p.y >= math.Min(a.y, b.y)-geometryEpsilon && p.y <= math.Max(a.y, b.y)+geometryEpsilon
}

func clipPolygonConvex(subject, clip []point) []point {
	if len(subject) < 3 || len(clip) < 3 {
		return nil
	}
	orientation := polygonArea(clip)
	result := subject
	for i := range clip {
		a, b := clip[i], clip[(i+1)%len(clip)]
		inside := func(p point) bool {
			value := cross(point{b.x - a.x, b.y - a.y}, point{p.x - a.x, p.y - a.y})
			if orientation >= 0 {
				return value >= -geometryEpsilon
			}
			return value <= geometryEpsilon
		}
		intersection := func(p, q point) point {
			t, _ := lineIntersection(p, q, a, b)
			return interpolate(p, q, t)
		}
		result = clipPolygonHalfPlane(result, inside, intersection)
	}
	return result
}

func clipPolygonHalfPlane(input []point, inside func(point) bool, intersection func(point, point) point) []point {
	if len(input) == 0 {
		return nil
	}
	var result []point
	previous := input[len(input)-1]
	previousInside := inside(previous)
	for _, current := range input {
		currentInside := inside(current)
		if currentInside != previousInside {
			result = append(result, intersection(previous, current))
		}
		if currentInside {
			result = append(result, current)
		}
		previous, previousInside = current, currentInside
	}
	return result
}

func intersectVertical(a, b point, x float64) point {
	if math.Abs(b.x-a.x) <= geometryEpsilon {
		return point{x, a.y}
	}
	t := (x - a.x) / (b.x - a.x)
	return interpolate(a, b, t)
}

func lineIntersection(p, p2, q, q2 point) (float64, float64) {
	r := point{p2.x - p.x, p2.y - p.y}
	s := point{q2.x - q.x, q2.y - q.y}
	denom := cross(r, s)
	if math.Abs(denom) <= geometryEpsilon {
		return 0, 0
	}
	qp := point{q.x - p.x, q.y - p.y}
	return cross(qp, s) / denom, cross(qp, r) / denom
}

func polygonArea(poly []point) float64 {
	area := 0.0
	for i := range poly {
		j := (i + 1) % len(poly)
		area += poly[i].x*poly[j].y - poly[j].x*poly[i].y
	}
	return area / 2
}

func isConvex(poly []point) bool {
	if len(poly) < 3 {
		return false
	}
	sign := 0.0
	for i := range poly {
		a, b, c := poly[i], poly[(i+1)%len(poly)], poly[(i+2)%len(poly)]
		value := cross(point{b.x - a.x, b.y - a.y}, point{c.x - b.x, c.y - b.y})
		if math.Abs(value) <= geometryEpsilon {
			continue
		}
		if sign == 0 {
			sign = value
		} else if sign*value < 0 {
			return false
		}
	}
	return true
}

func setLineBBoxFromPolygon(line *TextLine, poly []point) {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, p := range poly {
		minX, minY = math.Min(minX, p.x), math.Min(minY, p.y)
		maxX, maxY = math.Max(maxX, p.x), math.Max(maxY, p.y)
	}
	line.HPOS, line.VPOS = roundGeometry(minX), roundGeometry(minY)
	line.Width, line.Height = roundGeometry(maxX-minX), roundGeometry(maxY-minY)
}

func uniqueSorted(values []float64) []float64 {
	sort.Float64s(values)
	result := make([]float64, 0, len(values))
	for _, value := range values {
		value = math.Max(0, math.Min(1, value))
		if len(result) == 0 || math.Abs(value-result[len(result)-1]) > geometryEpsilon {
			result = append(result, value)
		}
	}
	return result
}

func interpolate(a, b point, t float64) point { return point{a.x + t*(b.x-a.x), a.y + t*(b.y-a.y)} }
func cross(a, b point) float64                { return a.x*b.y - a.y*b.x }
func samePoint(a, b point) bool               { return math.Hypot(a.x-b.x, a.y-b.y) <= geometryEpsilon }
func clampInt(v, low, high int) int           { return max(low, min(high, v)) }
func roundGeometry(v float64) float64         { return math.Round(v*1000) / 1000 }

func lineRuneCount(line TextLine) int {
	total := 0
	for _, token := range line.Strings {
		total += utf8.RuneCountInString(token.Content)
	}
	return total
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result[value] = true
		}
	}
	return result
}

func tagLabels(tagRefs string, labels map[string]string) []string {
	result := make([]string, 0)
	for _, ref := range strings.Fields(tagRefs) {
		if label, ok := labels[ref]; ok {
			result = append(result, label)
		}
	}
	return result
}

func intersectsSet(values []string, set map[string]bool) bool {
	for _, value := range values {
		if set[value] {
			return true
		}
	}
	return false
}

func discoverALTOPages(root string, flatOnly bool) (map[string]string, error) {
	result := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".xml") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if flatOnly && filepath.Dir(rel) != "." {
			return nil
		}
		key := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if strings.EqualFold(key, "original") {
			key = filepath.Base(filepath.Dir(path))
		}
		if previous, exists := result[key]; exists {
			return fmt.Errorf("multiple XML files map to page %q: %q and %q", key, previous, path)
		}
		result[key] = path
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("directory %q contains no matching XML files", root)
	}
	return result, nil
}

func samePath(a, b string) (bool, error) {
	absA, err := filepath.Abs(a)
	if err != nil {
		return false, fmt.Errorf("resolve path %q: %w", a, err)
	}
	absB, err := filepath.Abs(b)
	if err != nil {
		return false, fmt.Errorf("resolve path %q: %w", b, err)
	}
	return filepath.Clean(absA) == filepath.Clean(absB), nil
}
