//go:build !nogocv

package formatcov

import (
	"context"
	"fmt"
	"image"
	"image/color"
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
	denoiseAdaptiveBlockSize = 61 // must be odd
	denoiseAdaptiveC         = 14

	// Contour-based speckle filter.
	denoiseSpeckleMinArea   = 1.0
	denoiseSpeckleMaxArea   = 55.0 // wider: catch larger bleed-through clusters
	denoiseSpeckleMinFill   = 0.20
	denoiseSpeckleMinAspect = 0.30
	denoiseSpeckleMaxAspect = 3.0

	denoiseMaskColor  = uint8(255)
	denoiseMinWorkers = 2
	denoiseMaxWorkers = 24

	denoiseEnhanceNeighborMaxGray  = 238
	denoiseEnhanceNeighborStrength = 0.35
	denoiseEnhanceCoreGamma        = 1.35
	denoiseEnhanceNeighborMinRatio = 0.25
	denoiseMaskRefineMaxGray       = 250
	denoiseSupportLocalGrayMargin  = 18
	denoiseBlobBucketSize          = 12
	denoiseBlobAreaFraction        = 0.000008
	denoiseBlobMinGray             = 140
	denoiseBlobMaxGray             = 254

	denoiseReferenceMinDim = 1240.0
)

type denoiseParams struct {
	backgroundKernelSize  int
	morphOpenSize         int
	speckleKillOpenSize   int
	speckleKillCloseSize  int
	foregroundRepairSize  int
	foregroundHoleMaxArea int
	speckleMaxWidth       int
	speckleMaxHeight      int
	minBlobArea           int
	maskRefineSupportSize int
	enhanceSupportSize    int
	blobMinArea           int
}

type supportPixel struct {
	r       int
	c       int
	blended uint8
}

func paramsForImage(rows, cols int) denoiseParams {
	minDim := float64(minInt(rows, cols))
	scale := minDim / denoiseReferenceMinDim
	if scale < 0.5 {
		scale = 0.5
	}
	if scale > 3.0 {
		scale = 3.0
	}

	return denoiseParams{
		backgroundKernelSize:  scaledOdd(51, scale, 15),
		morphOpenSize:         scaledInt(1, scale, 1),
		speckleKillOpenSize:   scaledInt(2, scale, 1),
		speckleKillCloseSize:  scaledInt(2, scale, 1),
		foregroundRepairSize:  scaledOdd(5, scale, 3),
		foregroundHoleMaxArea: scaledArea(96, scale, 24),
		speckleMaxWidth:       scaledInt(12, scale, 4),
		speckleMaxHeight:      scaledInt(12, scale, 4),
		minBlobArea:           scaledArea(8, scale, 4),
		maskRefineSupportSize: scaledOdd(11, scale, 5),
		enhanceSupportSize:    scaledOdd(5, scale, 3),
		blobMinArea:          maxInt(1, int(math.Round(float64(rows*cols)*denoiseBlobAreaFraction))),
	}
}

func scaledInt(base int, scale float64, minVal int) int {
	v := int(math.Round(float64(base) * scale))
	if v < minVal {
		return minVal
	}
	return v
}

func scaledOdd(base int, scale float64, minVal int) int {
	v := scaledInt(base, scale, minVal)
	if v%2 == 0 {
		v++
	}
	return v
}

func scaledArea(base int, scale float64, minVal int) int {
	v := int(math.Round(float64(base) * scale * scale))
	if v < minVal {
		return minVal
	}
	return v
}

func maxDenoiseWorkers() int {
	n := runtime.NumCPU()
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

// denoiseOne pipeline:
//  1. Grayscale
//  2. Background normalization (morphological close / divide) — kills bleed-through & uneven illumination
//  3. Adaptive threshold — robust local binarization
//  4. Morphological open (small) — remove 1-2px noise strands
//  5. Second morph pass (larger open + close) — kill surviving diffuse speckles
//  6. Connected-component minimum-size filter — unconditionally drop tiny blobs
//  7. Contour shape filter — remove speckle-shaped blobs
//  8. Apply cleaned foreground mask to normalized grayscale output
func denoiseOne(inPath, outPath string) error {
	// --- 1. Load & grayscale ---
	img := gocv.IMRead(inPath, gocv.IMReadColor)
	if img.Empty() {
		return fmt.Errorf("read image %q: empty or unsupported format", inPath)
	}
	defer img.Close()

	gray := gocv.NewMat()
	defer gray.Close()
	if err := gocv.CvtColor(img, &gray, gocv.ColorBGRToGray); err != nil {
		return fmt.Errorf("convert to grayscale: %w", err)
	}
	params := paramsForImage(gray.Rows(), gray.Cols())

	// --- 2. Background normalization ---
	bgKernel := gocv.GetStructuringElement(
		gocv.MorphRect,
		image.Point{X: params.backgroundKernelSize, Y: params.backgroundKernelSize},
	)
	defer bgKernel.Close()

	background := gocv.NewMat()
	defer background.Close()
	gocv.MorphologyEx(gray, &background, gocv.MorphClose, bgKernel)

	grayF := gocv.NewMat()
	defer grayF.Close()
	gray.ConvertTo(&grayF, gocv.MatTypeCV32F)

	bgF := gocv.NewMat()
	defer bgF.Close()
	background.ConvertTo(&bgF, gocv.MatTypeCV32F)

	normalizedF := gocv.NewMat()
	defer normalizedF.Close()
	gocv.Divide(grayF, bgF, &normalizedF)

	normalized := gocv.NewMat()
	defer normalized.Close()
	gocv.ConvertScaleAbs(normalizedF, &normalized, 255.0, 0)

	// --- 2b. Median blur before thresholding ---
	// A 3x3 median blur kills isolated single/double pixel dots before binarization
	// without blurring text edges the way Gaussian blur would.
	blurred := gocv.NewMat()
	defer blurred.Close()
	gocv.MedianBlur(normalized, &blurred, 3)

	// --- 3. Adaptive threshold ---
	binary := gocv.NewMat()
	defer binary.Close()
	gocv.AdaptiveThreshold(
		blurred, &binary, 255,
		gocv.AdaptiveThresholdGaussian,
		gocv.ThresholdBinaryInv, // ink=white, background=black
		denoiseAdaptiveBlockSize,
		denoiseAdaptiveC,
	)

	// --- 4. First morphological open (small) ---
	k1 := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: params.morphOpenSize, Y: params.morphOpenSize},
	)
	defer k1.Close()

	opened1 := gocv.NewMat()
	defer opened1.Close()
	gocv.MorphologyEx(binary, &opened1, gocv.MorphOpen, k1)

	// --- 5. Second morph pass: larger open then close ---
	// The larger open erodes away isolated speckle clusters that survived step 4.
	// The subsequent close reconnects any text strokes that got slightly broken.
	k2open := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: params.speckleKillOpenSize, Y: params.speckleKillOpenSize},
	)
	defer k2open.Close()

	opened2 := gocv.NewMat()
	defer opened2.Close()
	gocv.MorphologyEx(opened1, &opened2, gocv.MorphOpen, k2open)

	k2close := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: params.speckleKillCloseSize, Y: params.speckleKillCloseSize},
	)
	defer k2close.Close()

	closed2 := gocv.NewMat()
	defer closed2.Close()
	gocv.MorphologyEx(opened2, &closed2, gocv.MorphClose, k2close)

	// --- 6. Connected-component minimum-size filter ---
	// Unconditionally removes every blob smaller than denoiseMinBlobArea pixels,
	// regardless of shape. This catches the finest residual speckle dust.
	afterCC, err := removeSmallBlobs(closed2, params.minBlobArea)
	if err != nil {
		return fmt.Errorf("remove small blobs: %w", err)
	}
	defer afterCC.Close()

	// --- 7. Contour shape filter ---
	// Removes blobs that are the right size to be speckles and have speckle-like
	// shape (near-square, moderately filled). Real text fragments tend to be
	// elongated or have low fill ratios.
	contours := gocv.FindContours(afterCC, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	speckleMask := gocv.Zeros(afterCC.Rows(), afterCC.Cols(), gocv.MatTypeCV8U)
	defer speckleMask.Close()

	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)
		if area < denoiseSpeckleMinArea || area > denoiseSpeckleMaxArea {
			continue
		}
		rect := gocv.BoundingRect(contour)
		w, h := rect.Dx(), rect.Dy()
		rectArea := w * h
		if rectArea <= 0 {
			continue
		}
		fillRatio := area / float64(rectArea)
		aspectRatio := float64(w) / float64(h)

		if w > params.speckleMaxWidth ||
			h > params.speckleMaxHeight ||
			fillRatio < denoiseSpeckleMinFill ||
			aspectRatio < denoiseSpeckleMinAspect ||
			aspectRatio > denoiseSpeckleMaxAspect {
			continue
		}

		if err := gocv.DrawContours(
			&speckleMask, contours, i,
			color.RGBA{R: denoiseMaskColor, G: denoiseMaskColor, B: denoiseMaskColor, A: denoiseMaskColor},
			-1,
		); err != nil {
			return fmt.Errorf("draw contour %d: %w", i, err)
		}
	}

	// Subtract speckle mask from the image.
	invSpeckleMask := gocv.NewMat()
	defer invSpeckleMask.Close()
	gocv.BitwiseNot(speckleMask, &invSpeckleMask)

	cleaned := gocv.NewMat()
	defer cleaned.Close()
	gocv.BitwiseAndWithMask(afterCC, afterCC, &cleaned, invSpeckleMask)

	repaired, err := repairForegroundMask(cleaned, params)
	if err != nil {
		return fmt.Errorf("repair foreground mask: %w", err)
	}
	defer repaired.Close()

	refined, err := refineForegroundMask(repaired, normalized, params)
	if err != nil {
		return fmt.Errorf("refine foreground mask: %w", err)
	}
	defer refined.Close()

	out, err := applyForegroundMask(normalized, refined, params)
	if err != nil {
		return fmt.Errorf("apply foreground mask: %w", err)
	}
	defer out.Close()

	if ok := gocv.IMWrite(outPath, out); !ok {
		return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
	}
	return nil
}

func applyForegroundMask(gray gocv.Mat, mask gocv.Mat, params denoiseParams) (gocv.Mat, error) {
	supportKernel := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: params.enhanceSupportSize, Y: params.enhanceSupportSize},
	)
	defer supportKernel.Close()

	support := gocv.NewMat()
	defer support.Close()
	gocv.Dilate(mask, &support, supportKernel)

	var supportPixels []supportPixel
	for r := 0; r < gray.Rows(); r++ {
		for c := 0; c < gray.Cols(); c++ {
			if support.GetUCharAt(r, c) == 0 || mask.GetUCharAt(r, c) != 0 {
				continue
			}
			if !hasEnoughMaskedNeighbors(mask, r, c) {
				continue
			}
			grayVal := gray.GetUCharAt(r, c)
			localMean, ok := localMaskedMeanGray(gray, mask, r, c)
			if !ok || int(grayVal) > localMean+denoiseSupportLocalGrayMargin {
				continue
			}
			enhanced := enhanceInk(grayVal)
			blended := blendGray(grayVal, enhanced, denoiseEnhanceNeighborStrength)
			supportPixels = append(supportPixels, supportPixel{r: r, c: c, blended: blended})
		}
	}

	out := gocv.NewMatWithSize(gray.Rows(), gray.Cols(), gocv.MatTypeCV8U)
	for r := 0; r < gray.Rows(); r++ {
		for c := 0; c < gray.Cols(); c++ {
			grayVal := gray.GetUCharAt(r, c)
			if mask.GetUCharAt(r, c) != 0 {
				out.SetUCharAt(r, c, enhanceInk(grayVal))
				continue
			}
			out.SetUCharAt(r, c, denoiseMaskColor)
		}
	}

	for _, px := range supportPixels {
		out.SetUCharAt(px.r, px.c, px.blended)
	}

	filteredOut, keptPixels, maxBlobArea := filterOutputByBlobSize(out, params.blobMinArea)
	log.Printf("formatcov output blobs: support_candidates=%d kept_pixels=%d min_blob_area=%d max_blob_area=%d gray_range=%d-%d bucket_size=%d", len(supportPixels), keptPixels, params.blobMinArea, maxBlobArea, denoiseBlobMinGray, denoiseBlobMaxGray, denoiseBlobBucketSize)
	out.Close()
	out = filteredOut

	return out, nil
}

func refineForegroundMask(mask gocv.Mat, tone gocv.Mat, params denoiseParams) (gocv.Mat, error) {
	kernel := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: params.maskRefineSupportSize, Y: params.maskRefineSupportSize},
	)
	defer kernel.Close()

	support := gocv.NewMat()
	defer support.Close()
	gocv.Dilate(mask, &support, kernel)

	refined := mask.Clone()
	for r := 0; r < tone.Rows(); r++ {
		for c := 0; c < tone.Cols(); c++ {
			if refined.GetUCharAt(r, c) != 0 {
				continue
			}
			if support.GetUCharAt(r, c) == 0 {
				continue
			}
			if tone.GetUCharAt(r, c) > denoiseMaskRefineMaxGray {
				continue
			}
			refined.SetUCharAt(r, c, denoiseMaskColor)
		}
	}

	return refined, nil
}

func maskNeighborCount(mask gocv.Mat, r, c int) int {
	count := 0
	r0 := maxInt(0, r-1)
	r1 := minInt(mask.Rows()-1, r+1)
	c0 := maxInt(0, c-1)
	c1 := minInt(mask.Cols()-1, c+1)
	for rr := r0; rr <= r1; rr++ {
		for cc := c0; cc <= c1; cc++ {
			if rr == r && cc == c {
				continue
			}
			if mask.GetUCharAt(rr, cc) != 0 {
				count++
			}
		}
	}
	return count
}

func hasEnoughMaskedNeighbors(mask gocv.Mat, r, c int) bool {
	r0 := maxInt(0, r-1)
	r1 := minInt(mask.Rows()-1, r+1)
	c0 := maxInt(0, c-1)
	c1 := minInt(mask.Cols()-1, c+1)

	possible := (r1-r0+1)*(c1-c0+1) - 1
	if possible <= 0 {
		return false
	}
	required := int(math.Ceil(float64(possible) * denoiseEnhanceNeighborMinRatio))
	if required < 1 {
		required = 1
	}
	return maskNeighborCount(mask, r, c) >= required
}

func repairForegroundMask(mask gocv.Mat, params denoiseParams) (gocv.Mat, error) {
	kernel := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: params.foregroundRepairSize, Y: params.foregroundRepairSize},
	)
	defer kernel.Close()

	repaired := gocv.NewMat()
	gocv.MorphologyEx(mask, &repaired, gocv.MorphClose, kernel)

	filled, err := fillSmallForegroundHoles(repaired, params.foregroundHoleMaxArea)
	repaired.Close()
	if err != nil {
		return gocv.NewMat(), err
	}
	return filled, nil
}

func fillSmallForegroundHoles(mask gocv.Mat, maxArea int) (gocv.Mat, error) {
	inv := gocv.NewMat()
	defer inv.Close()
	gocv.BitwiseNot(mask, &inv)

	labels := gocv.NewMat()
	defer labels.Close()
	stats := gocv.NewMat()
	defer stats.Close()
	centroids := gocv.NewMat()
	defer centroids.Close()

	n := gocv.ConnectedComponentsWithStats(inv, &labels, &stats, &centroids)
	fill := make([]bool, n)
	rows := mask.Rows()
	cols := mask.Cols()

	for label := 1; label < n; label++ {
		left := int(stats.GetIntAt(label, 0))
		top := int(stats.GetIntAt(label, 1))
		width := int(stats.GetIntAt(label, 2))
		height := int(stats.GetIntAt(label, 3))
		area := int(stats.GetIntAt(label, 4))

		touchesBorder := left == 0 || top == 0 || left+width >= cols || top+height >= rows
		if touchesBorder {
			continue
		}
		if area <= maxArea {
			fill[label] = true
		}
	}

	out := mask.Clone()
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			lbl := labels.GetIntAt(r, c)
			if int(lbl) < n && fill[lbl] {
				out.SetUCharAt(r, c, denoiseMaskColor)
			}
		}
	}

	return out, nil
}

func enhanceInk(v uint8) uint8 {
	x := float64(v) / 255.0
	if x <= 0 {
		return 0
	}
	enhanced := 255.0 * powFloat(x, denoiseEnhanceCoreGamma)
	if enhanced < 0 {
		enhanced = 0
	}
	if enhanced > 255 {
		enhanced = 255
	}
	return uint8(enhanced + 0.5)
}

func blendGray(base, target uint8, strength float64) uint8 {
	v := (1.0-strength)*float64(base) + strength*float64(target)
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return uint8(v + 0.5)
}

func localMaskedMeanGray(gray gocv.Mat, mask gocv.Mat, r, c int) (int, bool) {
	sum := 0
	count := 0
	r0 := maxInt(0, r-1)
	r1 := minInt(mask.Rows()-1, r+1)
	c0 := maxInt(0, c-1)
	c1 := minInt(mask.Cols()-1, c+1)
	for rr := r0; rr <= r1; rr++ {
		for cc := c0; cc <= c1; cc++ {
			if mask.GetUCharAt(rr, cc) == 0 {
				continue
			}
			sum += int(gray.GetUCharAt(rr, cc))
			count++
		}
	}
	if count == 0 {
		return 0, false
	}
	return sum / count, true
}

func keepSupportPixelsByBlobSize(rows, cols int, pixels []supportPixel, minArea int, bucketSize int) ([]bool, int) {
	keep := make([]bool, len(pixels))
	if len(pixels) == 0 {
		return keep, 0
	}

	grid := make([]int, rows*cols)
	for i := range grid {
		grid[i] = -1
	}
	for i, px := range pixels {
		grid[px.r*cols+px.c] = i
	}

	seen := make([]bool, len(pixels))
	queue := make([]int, 0, 256)
	component := make([]int, 0, 256)
	maxBlobArea := 0

	for i, px := range pixels {
		if seen[i] {
			continue
		}

		bucketLo := (int(px.blended) / bucketSize) * bucketSize
		bucketHi := bucketLo + bucketSize - 1

		queue = append(queue[:0], i)
		component = append(component[:0], i)
		seen[i] = true

		for head := 0; head < len(queue); head++ {
			currIdx := queue[head]
			curr := pixels[currIdx]
			r0 := maxInt(0, curr.r-1)
			r1 := minInt(rows-1, curr.r+1)
			c0 := maxInt(0, curr.c-1)
			c1 := minInt(cols-1, curr.c+1)
			for rr := r0; rr <= r1; rr++ {
				for cc := c0; cc <= c1; cc++ {
					neighborIdx := grid[rr*cols+cc]
					if neighborIdx < 0 || seen[neighborIdx] {
						continue
					}
					neighborGray := int(pixels[neighborIdx].blended)
					if neighborGray < bucketLo || neighborGray > bucketHi {
						continue
					}
					seen[neighborIdx] = true
					queue = append(queue, neighborIdx)
					component = append(component, neighborIdx)
				}
			}
		}

		if len(component) > maxBlobArea {
			maxBlobArea = len(component)
		}
		if len(component) >= minArea {
			for _, idx := range component {
				keep[idx] = true
			}
		}
	}

	return keep, maxBlobArea
}

func filterOutputByBlobSize(src gocv.Mat, minArea int) (gocv.Mat, int, int) {
	rows := src.Rows()
	cols := src.Cols()

	var pixels []supportPixel
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			v := src.GetUCharAt(r, c)
			if v >= denoiseMaskColor || v < denoiseBlobMinGray || v > denoiseBlobMaxGray {
				continue
			}
			pixels = append(pixels, supportPixel{r: r, c: c, blended: v})
		}
	}

	keep, maxBlobArea := keepSupportPixelsByBlobSize(rows, cols, pixels, minArea, denoiseBlobBucketSize)
	dst := src.Clone()
	kept := 0
	for i, px := range pixels {
		if keep[i] {
			kept++
			continue
		}
		dst.SetUCharAt(px.r, px.c, denoiseMaskColor)
	}

	return dst, kept, maxBlobArea
}

func powFloat(x, p float64) float64 {
	if p == 1 {
		return x
	}
	return math.Pow(x, p)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// removeSmallBlobs removes all connected components whose area in pixels is
// strictly less than minArea. Returns a new Mat (caller must Close it).
//
// Instead of iterating per-label over every pixel, we build a lookup table
// (label -> keep?) and then do a single pass over the labels image to build
// the output mask. This is O(width*height) rather than O(labels*width*height).
func removeSmallBlobs(src gocv.Mat, minArea int) (gocv.Mat, error) {
	labels := gocv.NewMat()
	defer labels.Close()
	stats := gocv.NewMat()
	defer stats.Close()
	centroids := gocv.NewMat()
	defer centroids.Close()

	n := gocv.ConnectedComponentsWithStats(src, &labels, &stats, &centroids)

	// Build a keep-table: keep[label] = true if area >= minArea.
	keep := make([]bool, n)
	for label := 1; label < n; label++ { // label 0 is background, always dropped
		area := stats.GetIntAt(label, 4) // col 4 = CC_STAT_AREA
		if int(area) >= minArea {
			keep[label] = true
		}
	}

	// Single pass: copy src pixel to dst only if its label is kept.
	dst := gocv.Zeros(src.Rows(), src.Cols(), gocv.MatTypeCV8U)
	for r := 0; r < src.Rows(); r++ {
		for c := 0; c < src.Cols(); c++ {
			lbl := labels.GetIntAt(r, c)
			if int(lbl) < n && keep[lbl] {
				dst.SetUCharAt(r, c, denoiseMaskColor)
			}
		}
	}

	return dst, nil
}
