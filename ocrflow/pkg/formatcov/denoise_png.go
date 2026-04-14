//go:build !nogocv

package formatcov

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gocv.io/x/gocv"
	"golang.org/x/sync/errgroup"
)

const (
	// denoiseMinWorkers bounds the denoise worker pool floor. Valid range: integer >= 1.
	denoiseMinWorkers = 1
	// denoiseMaxWorkers bounds the denoise worker pool ceiling. Valid range: integer >= denoiseMinWorkers.
	denoiseMaxWorkers = 2

	// denoiseProximityRatio scales the shared Chebyshev-radius proximity window from the smaller image dimension. Valid range: 0 < value < 1.
	denoiseProximityRatio = 0.01
	// denoiseSigmoidStrength controls how aggressively the S-curve pushes midtones toward black or white. Valid range: value > 0.
	denoiseSigmoidStrength = 14.0
	// denoiseT1DarkThresholdRatio marks t1 pixels darker than this grayscale level as foreground when detecting specks. Valid range: 0 <= value <= 1.
	denoiseT1DarkThresholdRatio = 0.92
	// denoiseSpeckInkThresholdRatio marks t1 pixels darker than this grayscale level as true speck ink so haze does not glue specks to nearby background. Valid range: 0 <= value <= 1.
	denoiseSpeckInkThresholdRatio = 0.82
	// denoiseSpeckIsolationThresholdRatio marks surrounding t1 pixels darker than this grayscale level as nearby dark support when deciding whether a speck is isolated. Valid range: 0 <= value <= 1.
	denoiseSpeckIsolationThresholdRatio = 0.58
	// denoiseT2DarknessRatio keeps only pixels with at least this normalized darkness in t2. Valid range: 0 <= value <= 1.
	denoiseT2DarknessRatio = 0.20
	// denoiseSpeckAreaRatio limits the maximum connected-component area eligible for speck removal, relative to full image area. Valid range: 0 < value < 1.
	denoiseSpeckAreaRatio = 0.00012
	// denoiseSpeckBBoxRatio limits the maximum bounding-box area eligible for speck removal, relative to full image area. Valid range: 0 < value < 1.
	denoiseSpeckBBoxRatio = 0.00045
	// denoiseSpeckDensityRatio requires a candidate speck to occupy at least this share of its own bounding box. Valid range: 0 < value <= 1.
	denoiseSpeckDensityRatio = 0.33
	// denoiseSpeckMeanDarknessRatio requires the average normalized darkness of a candidate speck to be at least this value before removal. Valid range: 0 <= value <= 1.
	denoiseSpeckMeanDarknessRatio = 0.35
	// denoiseSpeckHaloRatio expands the erased area around a removed speck by this share of the shared proximity radius so faint halos are cleared with the core blob. Valid range: 0 <= value <= 1.
	denoiseSpeckHaloRatio = 0.35
	// denoiseSpeckContextRatio limits how much broad t1 foreground may appear in the shared proximity neighborhood around a candidate speck before it is treated as page content instead. Valid range: 0 <= value <= 1.
	denoiseSpeckContextRatio = 0.015
	// denoiseBlankSpeckThresholdRatio marks merged-image pixels darker than this grayscale level as blank-area speck candidates in the final cleanup pass. Valid range: 0 <= value <= 1.
	denoiseBlankSpeckThresholdRatio = 0.92
	// denoiseBlankContextThresholdRatio marks merged-image pixels darker than this grayscale level as nearby content when deciding whether a final-pass speck sits in an otherwise blank area. Valid range: 0 <= value <= 1.
	denoiseBlankContextThresholdRatio = 0.70
	// denoiseBlankContextRatio limits how much nearby content may exist around a final-pass speck candidate before it is preserved as real page content. Valid range: 0 <= value <= 1.
	denoiseBlankContextRatio = 0.010
	// denoiseBlankHaloRatio expands the erased area around a final-pass blank-area speck by this share of the shared proximity radius so soft halos are cleared together with the core. Valid range: 0 <= value <= 1.
	denoiseBlankHaloRatio = 0.50
	// denoiseBlankSupportThresholdRatio marks t1 pixels darker than this grayscale level as real nearby content that protects a final-pass candidate from removal. Valid range: 0 <= value <= 1.
	denoiseBlankSupportThresholdRatio = 0.88
	// denoiseBlankSupportRadiusRatio scales the title/content protection radius for the final blank-area cleanup from the shared proximity radius. Valid range: value >= 1.
	denoiseBlankSupportRadiusRatio = 3.00
	// denoiseMergeWhiteThresholdRatio treats stage pixels at or above this grayscale level as white during the final merge so faint background haze does not win the darker-value comparison. Valid range: 0 <= value <= 1.
	denoiseMergeWhiteThresholdRatio = 0.90
	// denoiseMergeForegroundThresholdRatio requires at least one stage pixel to be darker than this grayscale level before the darker-value merge is allowed to keep non-white output. Valid range: 0 <= value <= 1.
	denoiseMergeForegroundThresholdRatio = 0.78
	// denoiseFadeMarginRatio scales the outward fade radius for gray-box edges from the smaller image dimension. Valid range: 0 < value < 1.
	denoiseFadeMarginRatio = 0.040
	// denoiseFadeGrayMinRatio is the minimum brightness level treated as gray-box source for the outward fade; pixels darker than this are ink. Valid range: 0 <= value <= 1.
	denoiseFadeGrayMinRatio = 0.39
)

func maxDenoiseWorkers() int {
	n := runtime.NumCPU() / 2
	if n < denoiseMinWorkers {
		return denoiseMinWorkers
	}
	if n > denoiseMaxWorkers {
		return denoiseMaxWorkers
	}
	return n
}

func DenoisePNGs(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read src dir %q: %w", src, err)
	}

	type job struct {
		inPath  string
		outPath string
	}
	var jobs []job
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.ToLower(filepath.Ext(name)) != ".png" {
			continue
		}
		jobs = append(jobs, job{
			inPath:  filepath.Join(src, name),
			outPath: filepath.Join(dst, name),
		})
	}

	if len(jobs) == 0 {
		return nil
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create dst dir %q: %w", dst, err)
	}

	workers := maxDenoiseWorkers()
	log.Printf("Denoising %d images with %d workers", len(jobs), workers)

	sem := make(chan struct{}, workers)
	grp, ctx := errgroup.WithContext(context.Background())

	for i, j := range jobs {
		i, j := i, j
		grp.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sem <- struct{}{}:
				defer func() { <-sem }()
			}
			log.Printf("[%d/%d] Denoising %q -> %q", i+1, len(jobs), j.inPath, j.outPath)
			if err := denoiseOne(j.inPath, j.outPath); err != nil {
				return fmt.Errorf("denoise %q: %w", j.inPath, err)
			}
			return nil
		})
	}

	return grp.Wait()
}

func denoiseOne(inPath, outPath string) error {
	img := gocv.IMRead(inPath, gocv.IMReadColor)
	if img.Empty() {
		return fmt.Errorf("read image %q: empty or unsupported format", inPath)
	}
	defer img.Close()

	gray, err := denoiseGray(img)
	if err != nil {
		return fmt.Errorf("convert %q to grayscale: %w", inPath, err)
	}
	defer gray.Close()

	t1, err := applyDenoiseSigmoid(gray)
	if err != nil {
		return fmt.Errorf("apply contrast curve to %q: %w", inPath, err)
	}
	defer t1.Close()

	proximityRadius := denoiseProximityRadius(gray.Rows(), gray.Cols())
	removedSpecks, err := removeIsolatedSpecks(&t1, proximityRadius)
	if err != nil {
		return fmt.Errorf("remove isolated specks from %q: %w", inPath, err)
	}

	t2, err := thresholdDarkPixels(gray)
	if err != nil {
		return fmt.Errorf("threshold dark pixels in %q: %w", inPath, err)
	}
	defer t2.Close()
	if err := removePixelsByMask(&t2, removedSpecks); err != nil {
		return fmt.Errorf("remove isolated specks from t2 in %q: %w", inPath, err)
	}

	if err := filterByProximity(&t2, t1, proximityRadius); err != nil {
		return fmt.Errorf("filter proximity-gated dark pixels in %q: %w", inPath, err)
	}

	combined, err := mergeDenoiseStages(t1, t2)
	if err != nil {
		return fmt.Errorf("merge denoise stages for %q: %w", inPath, err)
	}
	defer combined.Close()
	if err := removeBlankAreaSpecks(&combined, t1, proximityRadius); err != nil {
		return fmt.Errorf("remove blank-area specks from %q: %w", inPath, err)
	}

	fadeRadius := maxInt(5, int(math.Ceil(float64(minInt(combined.Rows(), combined.Cols()))*denoiseFadeMarginRatio)))
	if err := fadeGrayBoxMargins(&combined, gray, fadeRadius); err != nil {
		return fmt.Errorf("fade gray box margins for %q: %w", inPath, err)
	}

	if ok := gocv.IMWrite(outPath, combined); !ok {
		return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
	}
	return nil
}

func denoiseGray(src gocv.Mat) (gocv.Mat, error) {
	gray := gocv.NewMat()
	if err := gocv.CvtColor(src, &gray, gocv.ColorBGRToGray); err != nil {
		gray.Close()
		return gocv.NewMat(), err
	}
	return gray, nil
}

func applyDenoiseSigmoid(src gocv.Mat) (gocv.Mat, error) {
	lutData := make([]byte, 256)
	for i := range lutData {
		x := float64(i) / 255.0
		y := 1.0 / (1.0 + math.Exp(-denoiseSigmoidStrength*(x-0.5)))
		lutData[i] = uint8(math.Round(y * 255.0))
	}

	lut, err := gocv.NewMatFromBytes(1, len(lutData), gocv.MatTypeCV8U, lutData)
	if err != nil {
		return gocv.NewMat(), err
	}
	defer lut.Close()

	dst := gocv.NewMat()
	if err := gocv.LUT(src, lut, &dst); err != nil {
		dst.Close()
		return gocv.NewMat(), err
	}
	return dst, nil
}

func thresholdDarkPixels(gray gocv.Mat) (gocv.Mat, error) {
	threshold := uint8(math.Round((1.0 - denoiseT2DarknessRatio) * 255.0))
	src := gray.ToBytes()
	dst := make([]byte, len(src))
	for i, px := range src {
		if px <= threshold {
			dst[i] = px
			continue
		}
		dst[i] = 255
	}

	mat, err := gocv.NewMatFromBytes(gray.Rows(), gray.Cols(), gocv.MatTypeCV8U, dst)
	if err != nil {
		return gocv.NewMat(), err
	}
	return mat, nil
}

func filterByProximity(t2 *gocv.Mat, t1 gocv.Mat, proximityRadius int) error {
	t1Data := t1.ToBytes()
	t2Data := t2.ToBytes()
	rows, cols := t1.Rows(), t1.Cols()
	darkThreshold := uint8(math.Round(denoiseT1DarkThresholdRatio * 255.0))

	nearby := buildProximityMask(t1Data, rows, cols, proximityRadius, func(px uint8) bool {
		return px <= darkThreshold
	})

	for i, px := range t2Data {
		if px == 255 || nearby[i] {
			continue
		}
		t2Data[i] = 255
	}

	filtered, err := gocv.NewMatFromBytes(rows, cols, gocv.MatTypeCV8U, t2Data)
	if err != nil {
		return err
	}
	t2.Close()
	*t2 = filtered
	return nil
}

func mergeDenoiseStages(t1, t2 gocv.Mat) (gocv.Mat, error) {
	t1Data := t1.ToBytes()
	t2Data := t2.ToBytes()
	merged := make([]byte, len(t1Data))
	whiteThreshold := uint8(math.Round(denoiseMergeWhiteThresholdRatio * 255.0))
	foregroundThreshold := uint8(math.Round(denoiseMergeForegroundThresholdRatio * 255.0))

	for i, t1px := range t1Data {
		t2px := t2Data[i]
		if t1px >= whiteThreshold {
			t1px = 255
		}
		if t2px >= whiteThreshold {
			t2px = 255
		}
		if t2px == 255 {
			if t1px > foregroundThreshold {
				merged[i] = 255
				continue
			}
			merged[i] = t1px
			continue
		}
		if t1px == 255 {
			merged[i] = t2px
			continue
		}
		if t1px < t2px {
			merged[i] = t1px
			continue
		}
		merged[i] = t2px
	}

	mat, err := gocv.NewMatFromBytes(t1.Rows(), t1.Cols(), gocv.MatTypeCV8U, merged)
	if err != nil {
		return gocv.NewMat(), err
	}
	return mat, nil
}

func removeIsolatedSpecks(t1 *gocv.Mat, proximityRadius int) ([]bool, error) {
	rows, cols := t1.Rows(), t1.Cols()
	data := t1.ToBytes()
	darkThreshold := uint8(math.Round(denoiseSpeckInkThresholdRatio * 255.0))
	isolationThreshold := uint8(math.Round(denoiseSpeckIsolationThresholdRatio * 255.0))
	contextThreshold := uint8(math.Round(denoiseT1DarkThresholdRatio * 255.0))
	removed := make([]bool, len(data))

	visited := make([]bool, len(data))
	maxArea := maxInt(1, int(math.Ceil(float64(rows*cols)*denoiseSpeckAreaRatio)))
	maxBBoxArea := maxInt(1, int(math.Ceil(float64(rows*cols)*denoiseSpeckBBoxRatio)))
	haloRadius := maxInt(1, int(math.Ceil(float64(proximityRadius)*denoiseSpeckHaloRatio)))

	for idx, px := range data {
		if visited[idx] || px > darkThreshold {
			continue
		}

		component, bounds := collectDarkComponent(data, rows, cols, idx, darkThreshold, visited)
		area := len(component)
		bboxArea := (bounds.maxRow - bounds.minRow + 1) * (bounds.maxCol - bounds.minCol + 1)
		density := float64(area) / float64(bboxArea)
		meanDarkness := componentMeanDarkness(data, component)

		if area > maxArea || bboxArea > maxBBoxArea || density < denoiseSpeckDensityRatio || meanDarkness < denoiseSpeckMeanDarknessRatio {
			continue
		}
		if !componentIsIsolated(data, rows, cols, component, bounds, isolationThreshold, proximityRadius) {
			continue
		}
		if !componentHasSparseContext(data, rows, cols, component, bounds, contextThreshold, proximityRadius) {
			continue
		}

		removeSpeckCore(data, removed, component)
		expandRemovedSpeckMask(removed, rows, cols, bounds, haloRadius)
	}

	cleaned, err := gocv.NewMatFromBytes(rows, cols, gocv.MatTypeCV8U, data)
	if err != nil {
		return nil, err
	}
	t1.Close()
	*t1 = cleaned
	return removed, nil
}

func removeSpeckCore(data []byte, removed []bool, component []int) {
	for _, idx := range component {
		data[idx] = 255
		removed[idx] = true
	}
}

func expandRemovedSpeckMask(removed []bool, rows, cols int, bounds componentBounds, haloRadius int) {
	minRow := maxInt(0, bounds.minRow-haloRadius)
	maxRow := minInt(rows-1, bounds.maxRow+haloRadius)
	minCol := maxInt(0, bounds.minCol-haloRadius)
	maxCol := minInt(cols-1, bounds.maxCol+haloRadius)

	for row := minRow; row <= maxRow; row++ {
		base := row * cols
		for col := minCol; col <= maxCol; col++ {
			removed[base+col] = true
		}
	}
}

func removePixelsByMask(img *gocv.Mat, removed []bool) error {
	if len(removed) == 0 {
		return nil
	}

	data := img.ToBytes()
	for i, drop := range removed {
		if drop {
			data[i] = 255
		}
	}

	cleaned, err := gocv.NewMatFromBytes(img.Rows(), img.Cols(), gocv.MatTypeCV8U, data)
	if err != nil {
		return err
	}
	img.Close()
	*img = cleaned
	return nil
}

func removeBlankAreaSpecks(img *gocv.Mat, t1 gocv.Mat, proximityRadius int) error {
	rows, cols := img.Rows(), img.Cols()
	data := img.ToBytes()
	t1Data := t1.ToBytes()
	darkThreshold := uint8(math.Round(denoiseBlankSpeckThresholdRatio * 255.0))
	contextThreshold := uint8(math.Round(denoiseBlankContextThresholdRatio * 255.0))
	supportThreshold := uint8(math.Round(denoiseBlankSupportThresholdRatio * 255.0))

	visited := make([]bool, len(data))
	maxArea := maxInt(1, int(math.Ceil(float64(rows*cols)*denoiseSpeckAreaRatio)))
	maxBBoxArea := maxInt(1, int(math.Ceil(float64(rows*cols)*denoiseSpeckBBoxRatio)))
	haloRadius := maxInt(1, int(math.Ceil(float64(proximityRadius)*denoiseBlankHaloRatio)))
	supportRadius := maxInt(proximityRadius, int(math.Ceil(float64(proximityRadius)*denoiseBlankSupportRadiusRatio)))

	for idx, px := range data {
		if visited[idx] || px > darkThreshold {
			continue
		}

		component, bounds := collectDarkComponent(data, rows, cols, idx, darkThreshold, visited)
		area := len(component)
		bboxArea := (bounds.maxRow - bounds.minRow + 1) * (bounds.maxCol - bounds.minCol + 1)
		if area > maxArea || bboxArea > maxBBoxArea {
			continue
		}
		if !componentHasBlankContext(data, rows, cols, component, bounds, contextThreshold, proximityRadius) {
			continue
		}
		if componentHasNearbySupport(t1Data, rows, cols, bounds, supportThreshold, supportRadius) {
			continue
		}

		eraseBounds(data, rows, cols, bounds, haloRadius)
	}

	cleaned, err := gocv.NewMatFromBytes(rows, cols, gocv.MatTypeCV8U, data)
	if err != nil {
		return err
	}
	img.Close()
	*img = cleaned
	return nil
}

func componentHasNearbySupport(data []byte, rows, cols int, bounds componentBounds, supportThreshold uint8, supportRadius int) bool {
	minRow := maxInt(0, bounds.minRow-supportRadius)
	maxRow := minInt(rows-1, bounds.maxRow+supportRadius)
	minCol := maxInt(0, bounds.minCol-supportRadius)
	maxCol := minInt(cols-1, bounds.maxCol+supportRadius)

	for row := minRow; row <= maxRow; row++ {
		base := row * cols
		for col := minCol; col <= maxCol; col++ {
			if data[base+col] <= supportThreshold {
				return true
			}
		}
	}

	return false
}

func eraseBounds(data []byte, rows, cols int, bounds componentBounds, haloRadius int) {
	minRow := maxInt(0, bounds.minRow-haloRadius)
	maxRow := minInt(rows-1, bounds.maxRow+haloRadius)
	minCol := maxInt(0, bounds.minCol-haloRadius)
	maxCol := minInt(cols-1, bounds.maxCol+haloRadius)

	for row := minRow; row <= maxRow; row++ {
		base := row * cols
		for col := minCol; col <= maxCol; col++ {
			data[base+col] = 255
		}
	}
}

type componentBounds struct {
	minRow int
	maxRow int
	minCol int
	maxCol int
}

func collectDarkComponent(data []byte, rows, cols, start int, darkThreshold uint8, visited []bool) ([]int, componentBounds) {
	queue := []int{start}
	visited[start] = true
	component := make([]int, 0, 16)
	startRow, startCol := start/cols, start%cols
	bounds := componentBounds{
		minRow: startRow,
		maxRow: startRow,
		minCol: startCol,
		maxCol: startCol,
	}

	for len(queue) > 0 {
		idx := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		component = append(component, idx)

		row, col := idx/cols, idx%cols
		if row < bounds.minRow {
			bounds.minRow = row
		}
		if row > bounds.maxRow {
			bounds.maxRow = row
		}
		if col < bounds.minCol {
			bounds.minCol = col
		}
		if col > bounds.maxCol {
			bounds.maxCol = col
		}

		for dr := -1; dr <= 1; dr++ {
			for dc := -1; dc <= 1; dc++ {
				if dr == 0 && dc == 0 {
					continue
				}
				r := row + dr
				c := col + dc
				if r < 0 || r >= rows || c < 0 || c >= cols {
					continue
				}
				next := r*cols + c
				if visited[next] || data[next] > darkThreshold {
					continue
				}
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}

	return component, bounds
}

func componentMeanDarkness(data []byte, component []int) float64 {
	if len(component) == 0 {
		return 0
	}

	var darknessSum float64
	for _, idx := range component {
		darknessSum += 1.0 - float64(data[idx])/255.0
	}
	return darknessSum / float64(len(component))
}

func buildProximityMask(data []byte, rows, cols, proximityRadius int, include func(uint8) bool) []bool {
	mask := make([]bool, len(data))
	if proximityRadius < 0 {
		return mask
	}

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if !include(data[idx]) {
				continue
			}

			minRow := maxInt(0, row-proximityRadius)
			maxRow := minInt(rows-1, row+proximityRadius)
			minCol := maxInt(0, col-proximityRadius)
			maxCol := minInt(cols-1, col+proximityRadius)

			for r := minRow; r <= maxRow; r++ {
				base := r * cols
				for c := minCol; c <= maxCol; c++ {
					mask[base+c] = true
				}
			}
		}
	}

	return mask
}

func componentIsIsolated(data []byte, rows, cols int, component []int, bounds componentBounds, darkThreshold uint8, proximityRadius int) bool {
	componentSet := make(map[int]struct{}, len(component))
	for _, idx := range component {
		componentSet[idx] = struct{}{}
	}

	minRow := maxInt(0, bounds.minRow-proximityRadius)
	maxRow := minInt(rows-1, bounds.maxRow+proximityRadius)
	minCol := maxInt(0, bounds.minCol-proximityRadius)
	maxCol := minInt(cols-1, bounds.maxCol+proximityRadius)

	for row := minRow; row <= maxRow; row++ {
		for col := minCol; col <= maxCol; col++ {
			idx := row*cols + col
			if data[idx] > darkThreshold {
				continue
			}
			if _, ok := componentSet[idx]; ok {
				continue
			}
			if pixelInComponentProximity(row, col, cols, component, proximityRadius) {
				return false
			}
		}
	}

	return true
}

func componentHasSparseContext(data []byte, rows, cols int, component []int, bounds componentBounds, contextThreshold uint8, proximityRadius int) bool {
	componentSet := make(map[int]struct{}, len(component))
	for _, idx := range component {
		componentSet[idx] = struct{}{}
	}

	minRow := maxInt(0, bounds.minRow-proximityRadius)
	maxRow := minInt(rows-1, bounds.maxRow+proximityRadius)
	minCol := maxInt(0, bounds.minCol-proximityRadius)
	maxCol := minInt(cols-1, bounds.maxCol+proximityRadius)

	neighborArea := (maxRow - minRow + 1) * (maxCol - minCol + 1)
	maxContextPixels := int(math.Ceil(float64(neighborArea) * denoiseSpeckContextRatio))
	contextPixels := 0

	for row := minRow; row <= maxRow; row++ {
		base := row * cols
		for col := minCol; col <= maxCol; col++ {
			idx := base + col
			if _, ok := componentSet[idx]; ok {
				continue
			}
			if data[idx] <= contextThreshold {
				contextPixels++
				if contextPixels > maxContextPixels {
					return false
				}
			}
		}
	}

	return true
}

func componentHasBlankContext(data []byte, rows, cols int, component []int, bounds componentBounds, contextThreshold uint8, proximityRadius int) bool {
	componentSet := make(map[int]struct{}, len(component))
	for _, idx := range component {
		componentSet[idx] = struct{}{}
	}

	minRow := maxInt(0, bounds.minRow-proximityRadius)
	maxRow := minInt(rows-1, bounds.maxRow+proximityRadius)
	minCol := maxInt(0, bounds.minCol-proximityRadius)
	maxCol := minInt(cols-1, bounds.maxCol+proximityRadius)

	neighborArea := (maxRow - minRow + 1) * (maxCol - minCol + 1)
	maxContextPixels := int(math.Ceil(float64(neighborArea) * denoiseBlankContextRatio))
	contextPixels := 0

	for row := minRow; row <= maxRow; row++ {
		base := row * cols
		for col := minCol; col <= maxCol; col++ {
			idx := base + col
			if _, ok := componentSet[idx]; ok {
				continue
			}
			if data[idx] <= contextThreshold {
				contextPixels++
				if contextPixels > maxContextPixels {
					return false
				}
			}
		}
	}

	return true
}

func pixelInComponentProximity(row, col, cols int, component []int, proximityRadius int) bool {
	for _, idx := range component {
		componentRow, componentCol := idx/cols, idx%cols
		if maxInt(absInt(row-componentRow), absInt(col-componentCol)) <= proximityRadius {
			return true
		}
	}
	return false
}

func fadeGrayBoxMargins(img *gocv.Mat, orig gocv.Mat, fadeRadius int) error {
	rows, cols := img.Rows(), img.Cols()
	data := img.ToBytes()
	origData := orig.ToBytes()
	fadeGrayMin := uint8(math.Round(denoiseFadeGrayMinRatio * 255.0))

	srcMask := make([]byte, len(data))
	for i, px := range data {
		if px >= fadeGrayMin && px < 255 {
			srcMask[i] = 0
		} else {
			srcMask[i] = 255
		}
	}

	srcMat, err := gocv.NewMatFromBytes(rows, cols, gocv.MatTypeCV8U, srcMask)
	if err != nil {
		return err
	}
	defer srcMat.Close()

	distMat := gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV32F)
	defer distMat.Close()
	labels := gocv.NewMat()
	defer labels.Close()

	if err := gocv.DistanceTransform(srcMat, &distMat, &labels, gocv.DistL2, gocv.DistanceMask5, gocv.DistanceLabelPixel); err != nil {
		return err
	}

	distBytes := distMat.ToBytes()
	result := make([]byte, len(data))
	copy(result, data)

	fr := float64(fadeRadius)
	for i, px := range data {
		if px != 255 {
			continue
		}
		off := i * 4
		bits := uint32(distBytes[off]) | uint32(distBytes[off+1])<<8 | uint32(distBytes[off+2])<<16 | uint32(distBytes[off+3])<<24
		d := float64(math.Float32frombits(bits))
		if d >= fr {
			continue
		}
		origPx := float64(origData[i])
		t := d / fr
		s := t * t * t * (t*(t*6-15) + 10)
		result[i] = uint8(math.Round(origPx + s*(255-origPx)))
	}

	cleaned, err := gocv.NewMatFromBytes(rows, cols, gocv.MatTypeCV8U, result)
	if err != nil {
		return err
	}
	img.Close()
	*img = cleaned
	return nil
}

func denoiseProximityRadius(rows, cols int) int {
	minDim := rows
	if cols < minDim {
		minDim = cols
	}
	return maxInt(1, int(math.Ceil(float64(minDim)*denoiseProximityRatio)))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
