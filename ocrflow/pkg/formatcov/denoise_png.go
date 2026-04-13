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

// Denoise pipeline overview:
//
//  1. Grayscale + background normalization (morphological close / divide) — flattens
//     uneven illumination and bleed-through so the adaptive threshold sees a clean signal.
//  2. Median blur — kills isolated 1-2px dots before binarization without smearing edges.
//  3. Adaptive threshold — local Gaussian binarization; ink=white, background=black.
//     Low-quality images (min dim < denoiseSmallImageMinDim) use a lower C to capture weak ink.
//  4. Morphological open (small kernel) — erodes 1-2px noise strands from the binary mask.
//  5. Second morph pass: larger open then close — open kills surviving diffuse speckle;
//     close reconnects strokes that were slightly broken by the open.
//  6. Connected-component minimum-size filter — unconditionally drops every blob whose
//     pixel count is below the image-relative minimum, catching residual speckle dust.
//  7. Contour shape filter — removes blobs that are speckle-shaped (small, compact,
//     near-square, moderate fill). Blobs that belong to a spatial cluster of similarly-sized
//     neighbors are exempt: clustered small blobs are likely diagram dots or structured
//     decoration, not random noise. Skipped entirely for low-quality images.
//  8. Foreground mask repair — morphological close fills small gaps; tiny enclosed holes
//     are flood-filled to restore weak glyph interiors.
//  9. Mask refinement — dilates the mask slightly and pulls in nearby dark pixels that
//     fall within the normalized-gray threshold, recovering faint ink without broadly
//     reintroducing noise.
// 10. Output rendering — ink lightness is measured from the foreground mask histogram;
//     pages with uniformly light ink (p10/p25 ratio above threshold) receive a stronger
//     gamma boost so their glyphs render dark enough to survive downstream filtering.
//     Masked pixels are gamma-enhanced (with the ink-lightness boost applied during
//     rendering); a support fringe blends neighbor pixels at reduced strength; margin
//     glyphs are seeded separately. Final cleanup depends on image quality:
//     - Normal: blob connectivity filter over light-gray output pixels removes leftover
//       speckle clusters using image-relative area thresholds and tone buckets; the
//       tone-bucket width also widens for light-ink pages so glyph pixels connect.
//     - Low-quality: any non-mask pixel brighter than denoiseLowQualityMaxOutputGray
//       is wiped to white — simpler than the blob filter and more robust at small scale.
//
// All kernel sizes and area thresholds scale with the image's min dimension relative to
// denoiseReferenceMinDim, so no parameter is an absolute pixel count.
// This denoise flow was tuned on testdata/denoise examples.

const (
	denoiseDebug = false

	denoiseAdaptiveBlockSizeBase   = 61 // must be odd
	denoiseAdaptiveBlockSizeMin    = 11
	denoiseAdaptiveC               = 14
	denoiseAdaptiveCLowQuality     = 8
	denoiseSmallImageMinDim        = 800
	denoiseLowQualityMaxOutputGray = 210

	// Contour-based speckle filter.
	denoiseSpeckleMinArea     = 1.0
	denoiseSpeckleMaxAreaBase = 55.0 // wider: catch larger bleed-through clusters
	denoiseSpeckleMaxAreaMin  = 8.0
	denoiseSpeckleMinFill     = 0.20
	denoiseSpeckleMinAspect   = 0.30
	denoiseSpeckleMaxAspect   = 3.0

	denoiseMaskColor  = uint8(255)
	denoiseMinWorkers = 1
	denoiseMaxWorkers = 2

	denoiseEnhanceNeighborStrength     = 0.35
	denoiseEnhanceCoreGamma            = 1.35
	denoiseEnhanceNeighborMinRatio     = 0.25
	denoiseMaskRefineMaxGray           = 210
	denoiseMaskRefineMarginMaxGray     = 228
	denoiseSupportLocalGrayMargin      = 12
	denoiseSupportMarginGrayMargin     = 24
	denoiseMarginSeedMaxGray           = 196
	denoiseMarginSeedNeighbors         = 1
	denoiseMarginSeedPasses            = 3
	denoiseBlobBucketSize              = 6
	denoiseBlobAreaFraction            = 0.000012
	denoiseBlobMinGray                 = 140
	denoiseBlobMaxGray                 = 254
	denoiseBlobMaxMeanGray             = 198.0
	denoiseBlobLargeAreaFactor         = 3
	denoiseBlobLargeMaxMeanGray        = 188.0
	denoiseBlobHugeAreaFactor          = 6
	denoiseBlobHugeMaxMeanGray         = 180.0
	denoiseBlobDarkAnchorGray          = 165
	denoiseBlobDarkAnchorRatio         = 0.15
	denoiseBlobDarkAnchorMinPixelsBase = 4
	denoiseBlobDarkAnchorMinPixelsMin  = 1
	denoiseBlobEdgeMaxMeanGray         = 186.0
	denoiseBlobEdgeDarkAnchors         = 2
	denoiseBlobMarginMaxGray           = 214
	denoiseSideMarginFraction          = 0.18
	denoiseSideMarginMinPixels         = 96
	denoiseTopMarginFraction           = 0.14
	denoiseTopMarginMinPixels          = 72

	denoiseClusterSearchRadiusFraction = 0.06
	denoiseClusterMinNeighbors         = 3
	denoiseClusterAreaRatioMax         = 4.0

	denoiseInkPercentile      = 10
	denoiseInkDarkReference   = 110
	denoiseInkOffsetMax       = 40
	denoiseInkBucketScale     = 4
	denoiseInkBucketMax       = 120
	denoiseInkUniformityRatio = 0.84
	denoiseInkGammaBoost      = 1.5

	denoiseOriginalBlendAlpha = 0.45
	denoiseWeakPixelAlpha     = 0.2

	denoiseBlendProximityBase = 30
	denoiseBlendProximityMin  = 3

	denoiseReferenceMinDim = 1240.0

	denoiseMedianBlurBase        = 3
	denoiseMedianBlurMin         = 3
	denoiseBackgroundKernelBase  = 51
	denoiseBackgroundKernelMin   = 15
	denoiseMorphOpenBase         = 1
	denoiseMorphOpenMin          = 1
	denoiseSpeckleKillOpenBase   = 2
	denoiseSpeckleKillOpenMin    = 1
	denoiseSpeckleKillCloseBase  = 2
	denoiseSpeckleKillCloseMin   = 1
	denoiseForegroundRepairBase  = 5
	denoiseForegroundRepairMin   = 3
	denoiseForegroundHoleBase    = 96
	denoiseForegroundHoleMin     = 24
	denoiseSpeckleMaxWidthBase   = 12
	denoiseSpeckleMaxWidthMin    = 4
	denoiseSpeckleMaxHeightBase  = 12
	denoiseSpeckleMaxHeightMin   = 4
	denoiseMinBlobAreaBase       = 8
	denoiseMinBlobAreaMin        = 4
	denoiseMaskRefineSupportBase = 11
	denoiseMaskRefineSupportMin  = 5
	denoiseEnhanceSupportBase    = 5
	denoiseEnhanceSupportMin     = 3
)

type supportPixel struct {
	r       int
	c       int
	blended uint8
}

func denoiseScale(rows, cols int) float64 {
	minDim := float64(minInt(rows, cols))
	scale := minDim / denoiseReferenceMinDim
	if scale < 0.5 {
		scale = 0.5
	}
	if scale > 3.0 {
		scale = 3.0
	}
	return scale
}

func denoiseLowQuality(rows, cols int) bool {
	return float64(minInt(rows, cols)) < denoiseSmallImageMinDim
}

func denoiseAdaptiveCForImage(rows, cols int) int {
	if denoiseLowQuality(rows, cols) {
		return denoiseAdaptiveCLowQuality
	}
	return denoiseAdaptiveC
}

func denoiseBackgroundKernelSize(rows, cols int) int {
	return scaledOdd(denoiseBackgroundKernelBase, denoiseScale(rows, cols), denoiseBackgroundKernelMin)
}

func denoiseMorphOpenSize(rows, cols int) int {
	return scaledInt(denoiseMorphOpenBase, denoiseScale(rows, cols), denoiseMorphOpenMin)
}

func denoiseSpeckleKillOpenSize(rows, cols int) int {
	return scaledInt(denoiseSpeckleKillOpenBase, denoiseScale(rows, cols), denoiseSpeckleKillOpenMin)
}

func denoiseSpeckleKillCloseSize(rows, cols int) int {
	return scaledInt(denoiseSpeckleKillCloseBase, denoiseScale(rows, cols), denoiseSpeckleKillCloseMin)
}

func denoiseForegroundRepairSize(rows, cols int) int {
	return scaledOdd(denoiseForegroundRepairBase, denoiseScale(rows, cols), denoiseForegroundRepairMin)
}

func denoiseForegroundHoleMaxArea(rows, cols int) int {
	return scaledArea(denoiseForegroundHoleBase, denoiseScale(rows, cols), denoiseForegroundHoleMin)
}

func denoiseSpeckleMaxWidth(rows, cols int) int {
	return scaledInt(denoiseSpeckleMaxWidthBase, denoiseScale(rows, cols), denoiseSpeckleMaxWidthMin)
}

func denoiseSpeckleMaxHeight(rows, cols int) int {
	return scaledInt(denoiseSpeckleMaxHeightBase, denoiseScale(rows, cols), denoiseSpeckleMaxHeightMin)
}

func denoiseMinBlobArea(rows, cols int) int {
	return scaledArea(denoiseMinBlobAreaBase, denoiseScale(rows, cols), denoiseMinBlobAreaMin)
}

func denoiseMedianBlurSize(rows, cols int) int {
	return scaledOdd(denoiseMedianBlurBase, denoiseScale(rows, cols), denoiseMedianBlurMin)
}

func denoiseAdaptiveBlockSize(rows, cols int) int {
	return scaledOdd(denoiseAdaptiveBlockSizeBase, denoiseScale(rows, cols), denoiseAdaptiveBlockSizeMin)
}

func denoiseBlobDarkAnchorMinPixels(rows, cols int) int {
	return scaledInt(denoiseBlobDarkAnchorMinPixelsBase, denoiseScale(rows, cols), denoiseBlobDarkAnchorMinPixelsMin)
}

func denoiseSpeckleMaxArea(rows, cols int) float64 {
	scale := denoiseScale(rows, cols)
	v := denoiseSpeckleMaxAreaBase * scale * scale
	if v < denoiseSpeckleMaxAreaMin {
		return denoiseSpeckleMaxAreaMin
	}
	return v
}

func denoiseMaskRefineSupportSize(rows, cols int) int {
	return scaledOdd(denoiseMaskRefineSupportBase, denoiseScale(rows, cols), denoiseMaskRefineSupportMin)
}

func denoiseEnhanceSupportSize(rows, cols int) int {
	return scaledOdd(denoiseEnhanceSupportBase, denoiseScale(rows, cols), denoiseEnhanceSupportMin)
}

func denoiseOutputBlobMinArea(rows, cols int) int {
	return maxInt(1, int(math.Round(float64(rows*cols)*denoiseBlobAreaFraction)))
}

func denoiseSideMarginWidth(cols int) int {
	margin := int(math.Round(float64(cols) * denoiseSideMarginFraction))
	if margin < denoiseSideMarginMinPixels {
		return denoiseSideMarginMinPixels
	}
	maxMargin := cols / 3
	if margin > maxMargin {
		return maxMargin
	}
	return margin
}

func denoiseTopMarginHeight(rows int) int {
	margin := int(math.Round(float64(rows) * denoiseTopMarginFraction))
	if margin < denoiseTopMarginMinPixels {
		return denoiseTopMarginMinPixels
	}
	maxMargin := rows / 4
	if margin > maxMargin {
		return maxMargin
	}
	return margin
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

	gray := gocv.NewMat()
	defer gray.Close()
	if err := gocv.CvtColor(img, &gray, gocv.ColorBGRToGray); err != nil {
		return fmt.Errorf("convert to grayscale: %w", err)
	}
	bgKernel := gocv.GetStructuringElement(
		gocv.MorphRect,
		image.Point{X: denoiseBackgroundKernelSize(gray.Rows(), gray.Cols()), Y: denoiseBackgroundKernelSize(gray.Rows(), gray.Cols())},
	)
	defer bgKernel.Close()

	background := gocv.NewMat()
	defer background.Close()
	if err := gocv.MorphologyEx(gray, &background, gocv.MorphClose, bgKernel); err != nil {
		return err
	}

	grayF := gocv.NewMat()
	defer grayF.Close()

	if err := gray.ConvertTo(&grayF, gocv.MatTypeCV32F); err != nil {
		return err
	}

	bgF := gocv.NewMat()
	defer bgF.Close()
	if err := background.ConvertTo(&bgF, gocv.MatTypeCV32F); err != nil {
		return err
	}

	normalizedF := gocv.NewMat()
	defer normalizedF.Close()
	if err := gocv.Divide(grayF, bgF, &normalizedF); err != nil {
		return err
	}

	normalized := gocv.NewMat()
	defer normalized.Close()
	if err := gocv.ConvertScaleAbs(normalizedF, &normalized, 255.0, 0); err != nil {
		return err
	}

	blurred := gocv.NewMat()
	defer blurred.Close()
	if err := gocv.MedianBlur(normalized, &blurred, denoiseMedianBlurSize(gray.Rows(), gray.Cols())); err != nil {
		return err
	}

	binary := gocv.NewMat()
	defer binary.Close()
	if err := gocv.AdaptiveThreshold(
		blurred, &binary, 255,
		gocv.AdaptiveThresholdGaussian,
		gocv.ThresholdBinaryInv, // ink=white, background=black
		denoiseAdaptiveBlockSize(gray.Rows(), gray.Cols()),
		float32(denoiseAdaptiveCForImage(gray.Rows(), gray.Cols())),
	); err != nil {
		return err
	}

	k1 := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: denoiseMorphOpenSize(gray.Rows(), gray.Cols()), Y: denoiseMorphOpenSize(gray.Rows(), gray.Cols())},
	)
	defer k1.Close()

	opened1 := gocv.NewMat()
	defer opened1.Close()
	if err := gocv.MorphologyEx(binary, &opened1, gocv.MorphOpen, k1); err != nil {
		return err
	}

	k2open := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: denoiseSpeckleKillOpenSize(gray.Rows(), gray.Cols()), Y: denoiseSpeckleKillOpenSize(gray.Rows(), gray.Cols())},
	)
	defer k2open.Close()

	opened2 := gocv.NewMat()
	defer opened2.Close()
	if err := gocv.MorphologyEx(opened1, &opened2, gocv.MorphOpen, k2open); err != nil {
		return err
	}

	k2close := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: denoiseSpeckleKillCloseSize(gray.Rows(), gray.Cols()), Y: denoiseSpeckleKillCloseSize(gray.Rows(), gray.Cols())},
	)
	defer k2close.Close()

	closed2 := gocv.NewMat()
	defer closed2.Close()
	if err := gocv.MorphologyEx(opened2, &closed2, gocv.MorphClose, k2close); err != nil {
		return err
	}

	debugSaveMat(&binary, "01_binary")
	debugSaveMat(&closed2, "02_closed2")
	afterCC, err := removeSmallBlobs(closed2, denoiseMinBlobArea(gray.Rows(), gray.Cols()))
	if err != nil {
		return fmt.Errorf("remove small blobs: %w", err)
	}
	defer afterCC.Close()
	debugSaveMat(&afterCC, "03_afterCC")

	var cleaned gocv.Mat
	if denoiseLowQuality(gray.Rows(), gray.Cols()) {
		cleaned = afterCC.Clone()
	} else {
		contours := gocv.FindContours(afterCC, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		defer contours.Close()

		speckleMaxArea := denoiseSpeckleMaxArea(gray.Rows(), gray.Cols())
		speckleMaxW := denoiseSpeckleMaxWidth(gray.Rows(), gray.Cols())
		speckleMaxH := denoiseSpeckleMaxHeight(gray.Rows(), gray.Cols())

		type candidate struct {
			idx      int
			centroid image.Point
			area     float64
		}
		var candidates []candidate
		isCandidate := make([]bool, contours.Size())

		for i := 0; i < contours.Size(); i++ {
			contour := contours.At(i)
			area := gocv.ContourArea(contour)
			if area < denoiseSpeckleMinArea || area > speckleMaxArea {
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
			if w > speckleMaxW || h > speckleMaxH ||
				fillRatio < denoiseSpeckleMinFill ||
				aspectRatio < denoiseSpeckleMinAspect ||
				aspectRatio > denoiseSpeckleMaxAspect {
				continue
			}
			cx := rect.Min.X + w/2
			cy := rect.Min.Y + h/2
			candidates = append(candidates, candidate{idx: i, centroid: image.Point{X: cx, Y: cy}, area: area})
			isCandidate[i] = true
		}

		centroids := make([]image.Point, len(candidates))
		areas := make([]float64, len(candidates))
		for i, c := range candidates {
			centroids[i] = c.centroid
			areas[i] = c.area
		}
		inCluster := buildClusterMap(centroids, areas, afterCC.Rows(), afterCC.Cols())

		clusterSaved := 0
		for i, c := range candidates {
			if inCluster[i] {
				clusterSaved++
				isCandidate[c.idx] = false
			}
		}
		debugLogf("contour filter: candidates=%d cluster_saved=%d will_remove=%d", len(candidates), clusterSaved, len(candidates)-clusterSaved)

		speckleMask := gocv.Zeros(afterCC.Rows(), afterCC.Cols(), gocv.MatTypeCV8U)
		defer speckleMask.Close()

		for i := 0; i < contours.Size(); i++ {
			if !isCandidate[i] {
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

		invSpeckleMask := gocv.NewMat()
		defer invSpeckleMask.Close()
		if err := gocv.BitwiseNot(speckleMask, &invSpeckleMask); err != nil {
			return err
		}

		cleaned = gocv.NewMat()
		if err := gocv.BitwiseAndWithMask(afterCC, afterCC, &cleaned, invSpeckleMask); err != nil {
			return err
		}
	}
	defer cleaned.Close()

	repaired, err := repairForegroundMask(cleaned)
	if err != nil {
		return fmt.Errorf("repair foreground mask: %w", err)
	}
	defer repaired.Close()

	refined, err := refineForegroundMask(repaired, &normalized)
	if err != nil {
		return fmt.Errorf("refine foreground mask: %w", err)
	}
	defer refined.Close()

	out, err := applyForegroundMask(&normalized, refined, inPath)
	if err != nil {
		return fmt.Errorf("apply foreground mask: %w", err)
	}
	defer out.Close()

	merged := gocv.NewMat()
	defer merged.Close()
	if err := gocv.AddWeighted(out, denoiseOriginalBlendAlpha, gray, denoiseOriginalBlendAlpha, 0, &merged); err != nil {
		return fmt.Errorf("merge with original: %w", err)
	}

	proxSize := scaledInt(denoiseBlendProximityBase, denoiseScale(out.Rows(), out.Cols()), denoiseBlendProximityMin)
	proxSrc := gocv.NewMatWithSize(out.Rows(), out.Cols(), gocv.MatTypeCV8U)
	for r := 0; r < out.Rows(); r++ {
		for c := 0; c < out.Cols(); c++ {
			if out.GetUCharAt(r, c) < denoiseMaskColor {
				proxSrc.SetUCharAt(r, c, denoiseMaskColor)
			}
		}
	}
	proxKernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{X: proxSize, Y: proxSize})
	proxDilated := gocv.NewMat()
	if err := gocv.Dilate(proxSrc, &proxDilated, proxKernel); err != nil {
		proxSrc.Close()
		proxKernel.Close()
		return fmt.Errorf("proximity dilate: %w", err)
	}
	proxSrc.Close()
	proxKernel.Close()
	defer proxDilated.Close()

	for r := 0; r < merged.Rows(); r++ {
		for c := 0; c < merged.Cols(); c++ {
			if proxDilated.GetUCharAt(r, c) == 0 {
				merged.SetUCharAt(r, c, denoiseMaskColor)
			}
		}
	}

	weakThreshold := uint8(math.Round(255.0 * (1.0 - denoiseWeakPixelAlpha)))
	for r := 0; r < merged.Rows(); r++ {
		for c := 0; c < merged.Cols(); c++ {
			if merged.GetUCharAt(r, c) >= weakThreshold {
				merged.SetUCharAt(r, c, denoiseMaskColor)
			}
		}
	}

	if ok := gocv.IMWrite(outPath, merged); !ok {
		return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
	}
	return nil
}

func applyForegroundMask(gray *gocv.Mat, mask gocv.Mat, inPath string) (gocv.Mat, error) {
	supportKernel := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: denoiseEnhanceSupportSize(gray.Rows(), gray.Cols()), Y: denoiseEnhanceSupportSize(gray.Rows(), gray.Cols())},
	)
	defer supportKernel.Close()

	support := gocv.NewMat()
	defer support.Close()
	if err := gocv.Dilate(mask, &support, supportKernel); err != nil {
		return gocv.NewMat(), err
	}

	var supportPixels []supportPixel
	for r := 0; r < gray.Rows(); r++ {
		for c := 0; c < gray.Cols(); c++ {
			if support.GetUCharAt(r, c) == 0 || mask.GetUCharAt(r, c) != 0 {
				continue
			}
			marginPixel := isMarginPixel(gray.Rows(), gray.Cols(), r, c)
			neighborCount := maskNeighborCount(&mask, r, c)
			if !hasEnoughMaskedNeighbors(&mask, r, c) {
				if !marginPixel || neighborCount < 1 {
					continue
				}
			}
			grayVal := gray.GetUCharAt(r, c)
			localMean, ok := localMaskedMeanGray(gray, &mask, r, c)
			if !ok {
				continue
			}
			localGrayMargin := denoiseSupportLocalGrayMargin
			if marginPixel {
				localGrayMargin = denoiseSupportMarginGrayMargin
			}
			if int(grayVal) > localMean+localGrayMargin {
				continue
			}
			blended := renderSupportPixel(gray.Rows(), gray.Cols(), r, c, grayVal)
			supportPixels = append(supportPixels, supportPixel{r: r, c: c, blended: blended})
		}
	}

	inkOffset := inkLightnessOffset(gray, mask)
	inkGamma := 1.0 + (float64(inkOffset)/float64(denoiseInkOffsetMax))*denoiseInkGammaBoost

	out := gocv.NewMatWithSize(gray.Rows(), gray.Cols(), gocv.MatTypeCV8U)
	for r := 0; r < gray.Rows(); r++ {
		for c := 0; c < gray.Cols(); c++ {
			grayVal := gray.GetUCharAt(r, c)
			if mask.GetUCharAt(r, c) != 0 {
				v := renderForegroundPixel(gray.Rows(), gray.Cols(), r, c, grayVal)
				if inkOffset > 0 {
					x := float64(v) / 255.0
					if x > 0 {
						darkened := 255.0 * powFloat(x, inkGamma)
						if darkened > 255 {
							darkened = 255
						}
						v = uint8(darkened + 0.5)
					}
				}
				out.SetUCharAt(r, c, v)
				continue
			}
			out.SetUCharAt(r, c, denoiseMaskColor)
		}
	}

	for _, px := range supportPixels {
		out.SetUCharAt(px.r, px.c, px.blended)
	}

	seedMarginGlyphs(gray, &out)

	lowQuality := denoiseLowQuality(gray.Rows(), gray.Cols())
	adaptiveC := denoiseAdaptiveCForImage(gray.Rows(), gray.Cols())
	debugLogf("[%s] low_quality=%v adaptive_c=%d support_candidates=%d ink_offset=%d ink_gamma=%.3f", inPath, lowQuality, adaptiveC, len(supportPixels), inkOffset, inkGamma)
	if lowQuality {
		rows := out.Rows()
		cols := out.Cols()
		wiped := 0
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				if mask.GetUCharAt(r, c) != 0 {
					continue
				}
				v := out.GetUCharAt(r, c)
				if v < denoiseMaskColor && v > denoiseLowQualityMaxOutputGray {
					out.SetUCharAt(r, c, denoiseMaskColor)
					wiped++
				}
			}
		}
		debugLogf("[%s] low_quality brightness cutoff: wiped=%d threshold=%d", inPath, wiped, denoiseLowQualityMaxOutputGray)
	} else {
		minBlobArea := denoiseOutputBlobMinArea(gray.Rows(), gray.Cols())
		bucketSize := denoiseBlobBucketSize + inkOffset*denoiseInkBucketScale
		if bucketSize > denoiseInkBucketMax {
			bucketSize = denoiseInkBucketMax
		}
		keptPixels, maxBlobArea := filterOutputByBlobSize(&out, minBlobArea, inkOffset)
		debugLogf("[%s] output blobs: kept_pixels=%d min_blob_area=%d max_blob_area=%d gray_range=%d-%d bucket_size=%d", inPath, keptPixels, minBlobArea, maxBlobArea, denoiseBlobMinGray, denoiseBlobMaxGray, bucketSize)
	}

	return out, nil
}

func refineForegroundMask(mask gocv.Mat, tone *gocv.Mat) (gocv.Mat, error) {
	kernel := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: denoiseMaskRefineSupportSize(tone.Rows(), tone.Cols()), Y: denoiseMaskRefineSupportSize(tone.Rows(), tone.Cols())},
	)
	defer kernel.Close()

	support := gocv.NewMat()
	defer support.Close()
	if err := gocv.Dilate(mask, &support, kernel); err != nil {
		return gocv.Mat{}, err
	}

	refined := mask.Clone()
	for r := 0; r < tone.Rows(); r++ {
		for c := 0; c < tone.Cols(); c++ {
			if refined.GetUCharAt(r, c) != 0 {
				continue
			}
			if support.GetUCharAt(r, c) == 0 {
				continue
			}
			maxGray := uint8(denoiseMaskRefineMaxGray)
			if isMarginPixel(tone.Rows(), tone.Cols(), r, c) {
				maxGray = denoiseMaskRefineMarginMaxGray
			}
			if tone.GetUCharAt(r, c) > maxGray {
				continue
			}
			refined.SetUCharAt(r, c, denoiseMaskColor)
		}
	}

	return refined, nil
}

func maskNeighborCount(mask *gocv.Mat, r, c int) int {
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

func hasEnoughMaskedNeighbors(mask *gocv.Mat, r, c int) bool {
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

func repairForegroundMask(mask gocv.Mat) (gocv.Mat, error) {
	kernel := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: denoiseForegroundRepairSize(mask.Rows(), mask.Cols()), Y: denoiseForegroundRepairSize(mask.Rows(), mask.Cols())},
	)
	defer kernel.Close()

	repaired := gocv.NewMat()
	if err := gocv.MorphologyEx(mask, &repaired, gocv.MorphClose, kernel); err != nil {
		return gocv.Mat{}, err
	}

	filled, err := fillSmallForegroundHoles(repaired, denoiseForegroundHoleMaxArea(mask.Rows(), mask.Cols()))
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

func renderForegroundPixel(rows, cols, r, c int, grayVal uint8) uint8 {
	return enhanceInk(grayVal)
}

func renderSupportPixel(rows, cols, r, c int, grayVal uint8) uint8 {
	return blendGray(grayVal, enhanceInk(grayVal), denoiseEnhanceNeighborStrength)
}

func seedMarginGlyphs(gray *gocv.Mat, out *gocv.Mat) {
	rows := gray.Rows()
	cols := gray.Cols()
	for pass := 0; pass < denoiseMarginSeedPasses; pass++ {
		changed := false
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				if !isMarginPixel(rows, cols, r, c) || out.GetUCharAt(r, c) != denoiseMaskColor {
					continue
				}
				grayVal := gray.GetUCharAt(r, c)
				if grayVal > denoiseMarginSeedMaxGray {
					continue
				}
				if marginSeedNeighborCount(gray, out, r, c, denoiseMarginSeedMaxGray) < denoiseMarginSeedNeighbors {
					continue
				}
				out.SetUCharAt(r, c, grayVal)
				changed = true
			}
		}
		if !changed {
			return
		}
	}
}

func marginSeedNeighborCount(gray *gocv.Mat, out *gocv.Mat, r, c int, maxGray uint8) int {
	count := 0
	r0 := maxInt(0, r-1)
	r1 := minInt(gray.Rows()-1, r+1)
	c0 := maxInt(0, c-1)
	c1 := minInt(gray.Cols()-1, c+1)
	for rr := r0; rr <= r1; rr++ {
		for cc := c0; cc <= c1; cc++ {
			if rr == r && cc == c {
				continue
			}
			if out.GetUCharAt(rr, cc) != denoiseMaskColor || gray.GetUCharAt(rr, cc) <= maxGray {
				count++
			}
		}
	}
	return count
}

func localMaskedMeanGray(gray *gocv.Mat, mask *gocv.Mat, r, c int) (int, bool) {
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
	sideMarginWidth := denoiseSideMarginWidth(cols)
	topMarginHeight := denoiseTopMarginHeight(rows)

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
		sumGray := int(px.blended)
		darkAnchors := 0
		touchesPageMargin := px.c < sideMarginWidth || px.c >= cols-sideMarginWidth || px.r < topMarginHeight
		if int(px.blended) <= denoiseBlobDarkAnchorGray {
			darkAnchors++
		}

		for head := 0; head < len(queue); head++ {
			currIdx := queue[head]
			curr := pixels[currIdx]
			for _, step := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				rr := curr.r + step[0]
				cc := curr.c + step[1]
				if rr < 0 || rr >= rows || cc < 0 || cc >= cols {
					continue
				}
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
				sumGray += neighborGray
				if pixels[neighborIdx].c < sideMarginWidth || pixels[neighborIdx].c >= cols-sideMarginWidth || pixels[neighborIdx].r < topMarginHeight {
					touchesPageMargin = true
				}
				if neighborGray <= denoiseBlobDarkAnchorGray {
					darkAnchors++
				}
			}
		}

		area := len(component)
		if area > maxBlobArea {
			maxBlobArea = area
		}
		meanGray := float64(sumGray) / float64(area)
		if keepBlobComponent(area, minArea, meanGray, darkAnchors, rows, cols) || keepEdgeBlobComponent(area, meanGray, darkAnchors, touchesPageMargin) {
			for _, idx := range component {
				keep[idx] = true
			}
		}
	}

	return keep, maxBlobArea
}

func keepBlobComponent(area int, minArea int, meanGray float64, darkAnchors, rows, cols int) bool {
	if area < minArea {
		return false
	}
	if meanGray <= blobMaxMeanGrayForArea(area, minArea) {
		return true
	}
	return darkAnchors >= blobRequiredDarkAnchors(area, rows, cols)
}

func keepEdgeBlobComponent(area int, meanGray float64, darkAnchors int, touchesSideMargin bool) bool {
	if !touchesSideMargin {
		return false
	}
	if meanGray <= denoiseBlobEdgeMaxMeanGray {
		return true
	}
	if meanGray <= denoiseBlobMarginMaxGray && darkAnchors >= 1 {
		return true
	}
	return area >= denoiseBlobEdgeDarkAnchors && darkAnchors >= denoiseBlobEdgeDarkAnchors
}

func blobMaxMeanGrayForArea(area int, minArea int) float64 {
	if area >= minArea*denoiseBlobHugeAreaFactor {
		return denoiseBlobHugeMaxMeanGray
	}
	if area >= minArea*denoiseBlobLargeAreaFactor {
		return denoiseBlobLargeMaxMeanGray
	}
	return denoiseBlobMaxMeanGray
}

func blobRequiredDarkAnchors(area, rows, cols int) int {
	required := int(math.Ceil(float64(area) * denoiseBlobDarkAnchorRatio))
	minPixels := denoiseBlobDarkAnchorMinPixels(rows, cols)
	if required < minPixels {
		return minPixels
	}
	return required
}

func inkHistogramPercentile(hist *[256]int, total, pct int) int {
	target := total * pct / 100
	cumulative := 0
	for v := 0; v < 256; v++ {
		cumulative += hist[v]
		if cumulative >= target {
			return v
		}
	}
	return 255
}

func inkLightnessOffset(gray *gocv.Mat, mask gocv.Mat) int {
	var hist [256]int
	total := 0
	for r := 0; r < gray.Rows(); r++ {
		for c := 0; c < gray.Cols(); c++ {
			if mask.GetUCharAt(r, c) == 0 {
				continue
			}
			hist[gray.GetUCharAt(r, c)]++
			total++
		}
	}
	if total == 0 {
		return 0
	}
	p10 := inkHistogramPercentile(&hist, total, 10)
	p25 := inkHistogramPercentile(&hist, total, 25)
	p50 := inkHistogramPercentile(&hist, total, 50)
	debugLogf("ink histogram: p10=%d p25=%d p50=%d", p10, p25, p50)
	if p25 <= 0 || float64(p10)/float64(p25) < denoiseInkUniformityRatio {
		return 0
	}
	offset := p10 - denoiseInkDarkReference
	if offset < 0 {
		return 0
	}
	if offset > denoiseInkOffsetMax {
		return denoiseInkOffsetMax
	}
	return offset
}

func filterOutputByBlobSize(mat *gocv.Mat, minArea int, inkOffset int) (int, int) {
	rows := mat.Rows()
	cols := mat.Cols()

	var pixels []supportPixel
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			v := mat.GetUCharAt(r, c)
			if v >= denoiseMaskColor || v > denoiseBlobMaxGray {
				continue
			}
			if v < denoiseBlobMinGray && !isMarginPixel(rows, cols, r, c) {
				continue
			}
			pixels = append(pixels, supportPixel{r: r, c: c, blended: v})
		}
	}

	bucketSize := denoiseBlobBucketSize + inkOffset*denoiseInkBucketScale
	if bucketSize > denoiseInkBucketMax {
		bucketSize = denoiseInkBucketMax
	}
	keep, maxBlobArea := keepSupportPixelsByBlobSize(rows, cols, pixels, minArea, bucketSize)
	kept := 0
	for i, px := range pixels {
		if keep[i] {
			kept++
			continue
		}
		mat.SetUCharAt(px.r, px.c, denoiseMaskColor)
	}

	return kept, maxBlobArea
}

func buildClusterMap(centroids []image.Point, areas []float64, rows, cols int) []bool {
	n := len(centroids)
	inCluster := make([]bool, n)
	if n == 0 {
		return inCluster
	}

	minDim := float64(minInt(rows, cols))
	radius := minDim * denoiseClusterSearchRadiusFraction
	radiusSq := radius * radius

	cellSize := int(math.Ceil(radius))
	if cellSize < 1 {
		cellSize = 1
	}
	gridW := (cols + cellSize - 1) / cellSize
	gridH := (rows + cellSize - 1) / cellSize

	type cellKey struct{ gr, gc int }
	grid := make(map[cellKey][]int, n)
	for i, pt := range centroids {
		gr := pt.Y / cellSize
		gc := pt.X / cellSize
		grid[cellKey{gr, gc}] = append(grid[cellKey{gr, gc}], i)
	}

	neighborRadius := int(math.Ceil(radius/float64(cellSize))) + 1
	for i, pt := range centroids {
		gr0 := pt.Y / cellSize
		gc0 := pt.X / cellSize
		count := 0
		for dgr := -neighborRadius; dgr <= neighborRadius && count < denoiseClusterMinNeighbors; dgr++ {
			for dgc := -neighborRadius; dgc <= neighborRadius && count < denoiseClusterMinNeighbors; dgc++ {
				gr := gr0 + dgr
				gc := gc0 + dgc
				if gr < 0 || gr >= gridH || gc < 0 || gc >= gridW {
					continue
				}
				for _, j := range grid[cellKey{gr, gc}] {
					if j == i {
						continue
					}
					dx := float64(centroids[j].X - pt.X)
					dy := float64(centroids[j].Y - pt.Y)
					if dx*dx+dy*dy > radiusSq {
						continue
					}
					ratio := areas[i] / areas[j]
					if ratio < 1.0 {
						ratio = 1.0 / ratio
					}
					if ratio > denoiseClusterAreaRatioMax {
						continue
					}
					count++
					if count >= denoiseClusterMinNeighbors {
						break
					}
				}
			}
		}
		if count >= denoiseClusterMinNeighbors {
			inCluster[i] = true
		}
	}
	return inCluster
}

func isMarginPixel(rows, cols, r, c int) bool {
	sideMarginWidth := denoiseSideMarginWidth(cols)
	topMarginHeight := denoiseTopMarginHeight(rows)
	return c < sideMarginWidth || c >= cols-sideMarginWidth || r < topMarginHeight
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

func debugLogf(format string, args ...any) {
	if !denoiseDebug {
		return
	}
	log.Printf(format, args...)
}

func debugSaveMat(mat *gocv.Mat, label string) {
	if !denoiseDebug {
		return
	}
	path := filepath.Join("/tmp/formatcov_debug", label+".png")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	gocv.IMWrite(path, *mat)
}

func removeSmallBlobs(src gocv.Mat, minArea int) (gocv.Mat, error) {
	labels := gocv.NewMat()
	defer labels.Close()
	stats := gocv.NewMat()
	defer stats.Close()
	centroids := gocv.NewMat()
	defer centroids.Close()

	n := gocv.ConnectedComponentsWithStats(src, &labels, &stats, &centroids)

	keep := make([]bool, n)
	for label := 1; label < n; label++ {
		area := stats.GetIntAt(label, 4)
		if int(area) >= minArea {
			keep[label] = true
		}
	}

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
