//go:build !nogocv

package formatcov

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/samber/lo"
	"gocv.io/x/gocv"
	"golang.org/x/sync/errgroup"
)

var errWriteFailed = errors.New("gocv IMWrite failed")

const (
	deskewDebug = false

	opencvRemapMaxDim = math.MaxInt16 - 1
	deskewMarginRatio = 0.10
	deskewFocusRatio  = 0.80
	deskewFocusMinInk = 0.20

	deskewMinWorkers = 1
	deskewMaxWorkers = 24

	deskewDownscaleMax    = 1600
	deskewTrimBorder      = true
	deskewMaxRotate       = 4.0
	deskewLineAngleLimit  = 4.0
	deskewProjectionLimit = 6.0
	deskewAngleStep       = 0.25
	deskewMinRotate       = 0.15
	deskewBackground      = uint8(255)
)

func deskewLogf(format string, args ...any) {
	if !deskewDebug {
		return
	}
	log.Printf(format, args...)
}

func maxDeskewWorkers() int {
	n := runtime.NumCPU() / 2
	if n < deskewMinWorkers {
		return deskewMinWorkers
	}
	if n > deskewMaxWorkers {
		return deskewMaxWorkers
	}
	return n
}

func clampDeskewAngle(angle float64) float64 {
	if angle < -deskewMaxRotate {
		return -deskewMaxRotate
	}
	if angle > deskewMaxRotate {
		return deskewMaxRotate
	}
	return angle
}

func DeskewPNGs(src string, dst string) error {
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

	workers := maxDeskewWorkers()
	log.Printf("Deskewing %d images with %d workers (opencv_projection)", len(jobs), workers)

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

			log.Printf("[%d/%d] Deskewing %q -> %q", i+1, len(jobs), j.inPath, j.outPath)

			if err := deskewOneProjection(j.inPath, j.outPath); err != nil {
				return fmt.Errorf("deskew %q: %w", j.inPath, err)
			}
			return nil
		})
	}

	return grp.Wait()
}

func EstimateDeskewAnglePNG(inPath string) (float64, error) {
	img := gocv.IMRead(inPath, gocv.IMReadColor)
	if img.Empty() {
		return 0, fmt.Errorf("read image %q: empty or unsupported format", inPath)
	}
	defer img.Close()

	small := img.Clone()
	defer small.Close()

	if deskewDownscaleMax > 0 {
		var err error
		small, err = resizeMaxSide(small, deskewDownscaleMax)
		if err != nil {
			return 0, fmt.Errorf("downscale for angle estimation: %w", err)
		}
	}

	angle, err := estimateSkewProjection(small, inPath)
	if err != nil {
		return 0, err
	}
	return angle, nil
}

func DeskewPNGFileWithAngle(inPath, outPath string, angle float64) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create output dir %q: %w", filepath.Dir(outPath), err)
	}

	img := gocv.IMRead(inPath, gocv.IMReadColor)
	if img.Empty() {
		return fmt.Errorf("read image %q: empty or unsupported format", inPath)
	}
	defer img.Close()

	if math.Abs(angle) < deskewMinRotate {
		if ok := gocv.IMWrite(outPath, img); !ok {
			return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
		}
		return nil
	}

	correctionAngle := clampDeskewAngle(angle)
	if exceedsWarpAffineLimit(img, correctionAngle) {
		if ok := gocv.IMWrite(outPath, img); !ok {
			return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
		}
		return nil
	}

	rot, err := rotateKeepAll(img, correctionAngle, deskewBackground)
	if err != nil {
		return fmt.Errorf("rotate %q: %w", inPath, err)
	}
	defer rot.Close()

	if ok := gocv.IMWrite(outPath, rot); !ok {
		return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
	}
	return nil
}

func deskewOneProjection(inPath, outPath string) error {
	img := gocv.IMRead(inPath, gocv.IMReadColor)
	if img.Empty() {
		return fmt.Errorf("read image %q: empty or unsupported format", inPath)
	}
	defer img.Close()

	// Work on a downscaled copy for angle estimation.
	small := img.Clone()
	defer small.Close()

	if deskewDownscaleMax > 0 {
		var err error
		small, err = resizeMaxSide(small, deskewDownscaleMax)
		if err != nil {
			return fmt.Errorf("downscale for angle estimation: %w", err)
		}
		// When resized, resizeMaxSide closed the clone and returned a new Mat; defer above closes current small.
	}

	angle, err := estimateSkewProjection(small, inPath)
	if err != nil {
		return err
	}

	deskewLogf("[%s] estimated angle=%.3f deg (minRotate=%.2f)", inPath, angle, deskewMinRotate)

	if math.Abs(angle) < deskewMinRotate {
		deskewLogf("[%s] skipping rotation (angle below threshold)", inPath)
		if ok := gocv.IMWrite(outPath, img); !ok {
			return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
		}
		return nil
	}

	correctionAngle := clampDeskewAngle(angle)
	deskewLogf("[%s] applying correction angle=%.3f deg", inPath, correctionAngle)

	if exceedsWarpAffineLimit(img, correctionAngle) {
		log.Printf("Skipping deskew for %q: image dimensions exceed OpenCV remap limit during rotation", inPath)
		if ok := gocv.IMWrite(outPath, img); !ok {
			return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
		}
		return nil
	}

	rot, err := rotateKeepAll(img, correctionAngle, deskewBackground)
	if err != nil {
		return fmt.Errorf("rotate %q: %w", inPath, err)
	}
	defer rot.Close()

	if ok := gocv.IMWrite(outPath, rot); !ok {
		return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
	}
	return nil
}

func resizeMaxSide(src gocv.Mat, maxSide int) (gocv.Mat, error) {
	h, w := src.Rows(), src.Cols()
	m := h
	if w > m {
		m = w
	}
	if m <= maxSide {
		return src, nil
	}

	scale := float64(maxSide) / float64(m)
	newW := int(math.Round(float64(w) * scale))
	newH := int(math.Round(float64(h) * scale))
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := gocv.NewMat()
	err := gocv.Resize(src, &dst, image.Pt(newW, newH), 0, 0, gocv.InterpolationArea)
	if err != nil {
		return dst, err
	}
	src.Close()
	return dst, nil
}

func estimateSkewProjection(img gocv.Mat, inPath string) (float64, error) {
	gray := gocv.NewMat()
	defer gray.Close()
	if err := gocv.CvtColor(img, &gray, gocv.ColorBGRToGray); err != nil {
		return 0, fmt.Errorf("convert to grayscale: %w", err)
	}

	if crop, ok := cropDeskewMargins(gray); ok {
		defer crop.Close()

		tmp := crop.Clone()
		gray.Close()
		gray = tmp
	}

	// Optional trim: crop to content bbox before scoring
	if deskewTrimBorder {
		var err error
		bin, err := adaptiveBinary(gray)
		if err != nil {
			return 0, fmt.Errorf("adaptive binary for content bbox: %w", err)
		}
		defer bin.Close()

		if rect, ok := contentBBox(bin); ok {
			crop := gray.Region(rect)
			defer crop.Close()

			tmp := crop.Clone()
			gray.Close()
			gray = tmp
		}
	}

	// Prepare inverted binary (ink=255, bg=0) once
	bin, err := adaptiveBinary(gray)
	if err != nil {
		return 0, fmt.Errorf("adaptive binary: %w", err)
	}
	defer bin.Close()

	inv := gocv.NewMat()
	defer inv.Close()
	if err := gocv.BitwiseNot(bin, &inv); err != nil {
		return 0, fmt.Errorf("invert binary: %w", err)
	}

	textMask, textMaskOK, err := buildDeskewTextMask(inv)
	if err != nil {
		return 0, err
	}
	if textMaskOK {
		defer textMask.Close()
	} else {
		textMask = inv
	}
	refineMask := textMask
	if focusMask, ok := centralDeskewFocusMask(textMask); ok {
		defer focusMask.Close()
		refineMask = focusMask
	}

	lineEvidence, err := estimateSkewTextLines(textMask, inPath)
	if err != nil {
		return 0, err
	}
	if !lineEvidence.ok {
		if ruledAngle, ok, err := estimateSparseRuledSkew(inv, inPath); err != nil {
			return 0, err
		} else if ok {
			return clampDeskewAngle(ruledAngle), nil
		}
	}
	lineAngle := lineEvidence.angle
	lineDispersion := lineEvidence.dispersion
	lineOK := lineEvidence.ok
	constrainProjectionSearch := lineOK
	allowNearZeroPlateau := !lineOK
	if lineOK {
		if lineAngle < -deskewLineAngleLimit {
			lineAngle = -deskewLineAngleLimit
		}
		if lineAngle > deskewLineAngleLimit {
			lineAngle = deskewLineAngleLimit
		}
		deskewLogf("[%s] line-angle angle=%.3f dispersion=%.3f", inPath, lineAngle, lineDispersion)
		if lineDispersion <= 0.75 {
			refineWindow := 0.75
			refineTolerance := 0.75
			if math.Abs(lineAngle) >= deskewLineAngleLimit-0.15 {
				refineWindow = 2.0
				refineTolerance = 2.0
				deskewLogf("[%s] line-angle near limit; widening local refine window to %.2f", inPath, refineWindow)
			}
			refineInput := textMask
			if math.Abs(lineAngle) <= 1.5 {
				refineInput = refineMask
			}
			refinedAngle, refinedScore, zeroScore, scoreSpan, err := refineProjectionAroundAngle(refineInput, lineAngle, refineWindow, inPath)
			if err != nil {
				return 0, err
			}
			if zeroScore > 0 && refinedScore <= zeroScore*0.995 {
				constrainProjectionSearch = false
				allowNearZeroPlateau = true
				deskewLogf("[%s] line-angle local projection not better than zero; widening fallback search", inPath)
			}
			if dampedAngle, ok := dampHoughOnlyAngle(lineAngle, lineDispersion, refinedAngle, scoreSpan, zeroScore, inPath, lineOK, lineEvidence.houghOnly, textMask); ok {
				return clampDeskewAngle(dampedAngle), nil
			}
			deskewLogf("[%s] line-angle refine candidate=%.3f score=%.6f zeroScore=%.6f scoreSpan=%.6f", inPath, refinedAngle, refinedScore, zeroScore, scoreSpan)
			if weakLineAngleEvidence(lineAngle, lineDispersion, refinedScore, zeroScore, scoreSpan) {
				deskewLogf("[%s] line-angle rejected as low-confidence angle=%.3f zeroScore=%.6f score=%.6f span=%.6f", inPath, lineAngle, zeroScore, refinedScore, scoreSpan)
				return 0, nil
			}
			if math.Abs(refinedAngle-lineAngle) <= 0.5 && lineDispersion <= 0.15 && projectionFlatPeak(refinedScore, zeroScore, scoreSpan) {
				deskewLogf("[%s] line-angle retained on shallow local peak angle=%.3f", inPath, lineAngle)
				return clampDeskewAngle(lineAngle), nil
			}
			if refinedAngle*lineAngle < 0 &&
				math.Max(math.Abs(lineAngle), math.Abs(refinedAngle)) <= 0.45 &&
				projectionFlatPeak(refinedScore, zeroScore, scoreSpan) {
				deskewLogf("[%s] line-angle rejected due to tiny flat sign-flipping refine line=%.3f refined=%.3f", inPath, lineAngle, refinedAngle)
				return 0, nil
			}
			if math.Abs(lineAngle) <= 1.0 &&
				refinedAngle*lineAngle > 0 &&
				math.Abs(refinedAngle-lineAngle) <= 0.35 &&
				refinedScore > zeroScore*1.01 {
				deskewLogf("[%s] line-angle accepted on same-sign small-angle refine line=%.3f refined=%.3f", inPath, lineAngle, refinedAngle)
				return clampDeskewAngle(refinedAngle), nil
			}
			if math.Abs(lineAngle) <= 1.0 && math.Abs(refinedAngle-lineAngle) > 0.5 && math.Abs(refinedAngle) > math.Abs(lineAngle) {
				deskewLogf("[%s] line-angle retained to avoid over-refining small tilt angle=%.3f refined=%.3f", inPath, lineAngle, refinedAngle)
				return clampDeskewAngle(lineAngle), nil
			}
			if math.Abs(refinedAngle-lineAngle) <= refineTolerance && projectionImprovementOK(refinedScore, zeroScore, refinedAngle) {
				return clampDeskewAngle(refinedAngle), nil
			}
			if math.Abs(refinedAngle-lineAngle) <= math.Max(1.0, refineTolerance) && projectionFlatPeak(refinedScore, zeroScore, scoreSpan) {
				deskewLogf("[%s] line-angle accepted on flat projection peak angle=%.3f", inPath, lineAngle)
				return clampDeskewAngle(lineAngle), nil
			}
			if projectionImprovementOK(refinedScore, zeroScore, refinedAngle) && scoreSpan <= zeroScore*0.01 {
				blended := 0.7*lineAngle + 0.3*refinedAngle
				if blended < -deskewLineAngleLimit {
					blended = -deskewLineAngleLimit
				}
				if blended > deskewLineAngleLimit {
					blended = deskewLineAngleLimit
				}
				deskewLogf("[%s] line-angle accepted on projection plateau blended=%.3f", inPath, blended)
				return clampDeskewAngle(blended), nil
			}
			if math.Abs(lineAngle) >= 0.75 && math.Abs(refinedAngle) <= 0.35 && math.Abs(refinedAngle-lineAngle) >= 0.5 {
				constrainProjectionSearch = false
				allowNearZeroPlateau = true
				deskewLogf("[%s] line-angle conflicts with near-zero projection shoulder; widening fallback search", inPath)
			}
			allowNearZeroPlateau = true
			deskewLogf("[%s] line-angle rejected by projection consistency", inPath)
		}
	}

	bestA := 0.0
	bestS := -1.0
	zeroScore := -1.0
	scoreSamples := make([]weightedAngleSample, 0, int(math.Ceil((2*deskewProjectionLimit)/deskewAngleStep))+1)
	searchMin := -deskewProjectionLimit
	searchMax := deskewProjectionLimit
	if constrainProjectionSearch {
		window := math.Max(1.5, 2*lineDispersion)
		searchMin = math.Max(-deskewProjectionLimit, lineAngle-window)
		searchMax = math.Min(deskewProjectionLimit, lineAngle+window)
	}

	for a := searchMin; a <= searchMax+1e-9; a += deskewAngleStep {
		rot, err := rotateSameSize(textMask, a)
		if err != nil {
			return 0, fmt.Errorf("rotate for angle %.2f: %w", a, err)
		}
		s := projectionVariance(rot)
		rot.Close()

		deskewLogf("[%s] angle=%.2f score=%.6f", inPath, a, s)
		scoreSamples = append(scoreSamples, weightedAngleSample{
			angle:  a,
			weight: s,
		})

		if s > bestS {
			bestS = s
			bestA = a
		}
		if math.Abs(a) <= deskewAngleStep/2 {
			zeroScore = s
		}
	}

	deskewLogf("[%s] best angle=%.3f score=%.6f", inPath, bestA, bestS)
	if zeroScore < 0 {
		rot0, err := rotateSameSize(textMask, 0)
		if err != nil {
			return 0, fmt.Errorf("rotate for zero score: %w", err)
		}
		zeroScore = projectionVariance(rot0)
		rot0.Close()
	}
	if !projectionImprovementOK(bestS, zeroScore, bestA) {
		if allowNearZeroPlateau && math.Abs(bestA) <= 0.5 {
			if biasedAngle, ok := biasedNearZeroPlateauAngle(scoreSamples, zeroScore); ok {
				deskewLogf("[%s] best angle accepted on biased near-zero plateau angle=%.3f", inPath, biasedAngle)
				return clampDeskewAngle(biasedAngle), nil
			}
		}
		deskewLogf("[%s] best angle rejected due to weak improvement best=%.6f zero=%.6f", inPath, bestS, zeroScore)
		return 0, nil
	}
	if !constrainProjectionSearch && math.Abs(bestA) >= 3.0 && (zeroScore <= 0 || bestS <= zeroScore*4.0) {
		deskewLogf("[%s] best angle rejected as unsupported projection-only tilt angle=%.3f best=%.6f zero=%.6f", inPath, bestA, bestS, zeroScore)
		return 0, nil
	}
	if projectionPeakAtBoundary(bestA, searchMin, searchMax) {
		deskewLogf("[%s] best angle rejected at search boundary angle=%.3f range=[%.2f, %.2f]", inPath, bestA, searchMin, searchMax)
		return 0, nil
	}

	return clampDeskewAngle(bestA), nil
}

func centralDeskewFocusMask(src gocv.Mat) (gocv.Mat, bool) {
	rows := src.Rows()
	cols := src.Cols()

	marginX := int(math.Round(float64(cols) * (1 - deskewFocusRatio) / 2))
	marginY := int(math.Round(float64(rows) * (1 - deskewFocusRatio) / 2))
	if marginX <= 0 && marginY <= 0 {
		return gocv.Mat{}, false
	}
	left := marginX
	top := marginY
	right := cols - marginX
	bottom := rows - marginY
	if left >= right || top >= bottom {
		return gocv.Mat{}, false
	}

	region := src.Region(image.Rect(left, top, right, bottom))
	defer region.Close()
	totalInk := gocv.CountNonZero(src)
	if totalInk == 0 {
		return gocv.Mat{}, false
	}
	if float64(gocv.CountNonZero(region))/float64(totalInk) < deskewFocusMinInk {
		return gocv.Mat{}, false
	}
	return region.Clone(), true
}

func cropDeskewMargins(src gocv.Mat) (gocv.Mat, bool) {
	rows := src.Rows()
	cols := src.Cols()
	if rows < 4 || cols < 4 {
		return gocv.Mat{}, false
	}

	marginX := int(math.Round(float64(cols) * deskewMarginRatio))
	marginY := int(math.Round(float64(rows) * deskewMarginRatio))
	if marginX <= 0 && marginY <= 0 {
		return gocv.Mat{}, false
	}

	left := marginX
	top := marginY
	right := cols - marginX
	bottom := rows - marginY
	if left >= right || top >= bottom {
		return gocv.Mat{}, false
	}

	return src.Region(image.Rect(left, top, right, bottom)), true
}

func refineProjectionAroundAngle(inv gocv.Mat, center float64, window float64, inPath string) (float64, float64, float64, float64, error) {
	if window <= 0 {
		window = 0.75
	}
	start := math.Max(-deskewProjectionLimit, center-window)
	end := math.Min(deskewProjectionLimit, center+window)

	bestA := 0.0
	bestS := -1.0
	minS := math.MaxFloat64
	for a := start; a <= end+1e-9; a += deskewAngleStep {
		rot, err := rotateSameSize(inv, a)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("rotate for local refine angle %.2f: %w", a, err)
		}
		s := projectionVariance(rot)
		rot.Close()
		deskewLogf("[%s] local angle=%.2f score=%.6f", inPath, a, s)
		if s > bestS {
			bestS = s
			bestA = a
		}
		if s < minS {
			minS = s
		}
	}

	rot0, err := rotateSameSize(inv, 0)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("rotate for zero score: %w", err)
	}
	zeroScore := projectionVariance(rot0)
	rot0.Close()

	return bestA, bestS, zeroScore, bestS - minS, nil
}

func projectionImprovementOK(bestScore, zeroScore, angle float64) bool {
	if bestScore <= 0 {
		return false
	}
	if zeroScore <= 0 {
		return true
	}
	requiredImprovement := 0.03 + 0.06*math.Max(0, math.Abs(angle)-1.5)
	return (bestScore-zeroScore)/zeroScore >= requiredImprovement
}

func projectionFlatPeak(bestScore, zeroScore, scoreSpan float64) bool {
	if bestScore <= 0 || zeroScore <= 0 {
		return false
	}
	improvement := (bestScore - zeroScore) / zeroScore
	spanRatio := scoreSpan / zeroScore
	return improvement >= 0.02 && improvement <= 0.15 && spanRatio <= 0.15
}

func projectionPeakAtBoundary(bestAngle, searchMin, searchMax float64) bool {
	margin := deskewAngleStep / 2
	return bestAngle <= searchMin+margin || bestAngle >= searchMax-margin
}

func biasedNearZeroPlateauAngle(samples []weightedAngleSample, zeroScore float64) (float64, bool) {
	if len(samples) == 0 || zeroScore <= 0 {
		return 0, false
	}

	scoreByAngle := make(map[int]float64, len(samples))
	for _, sample := range samples {
		key := int(math.Round(sample.angle / deskewAngleStep))
		scoreByAngle[key] = sample.weight
	}

	pairedBias := 0.0
	pairedWeight := 0.0
	for angle := 0.50; angle <= 1.5+1e-9; angle += deskewAngleStep {
		key := int(math.Round(angle / deskewAngleStep))
		posScore, okPos := scoreByAngle[key]
		negScore, okNeg := scoreByAngle[-key]
		if !okPos || !okNeg {
			continue
		}
		diff := negScore - posScore
		weight := 1.75 - angle
		if weight < 0.2 {
			weight = 0.2
		}
		pairedBias += diff * weight
		pairedWeight += math.Abs(diff) * weight
	}

	if pairedWeight == 0 {
		return 0, false
	}
	normalizedBias := math.Abs(pairedBias) / zeroScore
	if normalizedBias < 0.01 {
		return 0, false
	}

	sign := 1.0
	if pairedBias > 0 {
		sign = -1.0
	}

	bestAngle := 0.0
	bestDistance := 0.0
	bestScore := 0.0
	for _, angle := range []float64{1.50, 1.25, 1.00, 0.75, 0.50, 0.25} {
		key := int(math.Round(angle / deskewAngleStep))
		score, ok := scoreByAngle[int(sign)*key]
		if !ok {
			continue
		}
		if score < zeroScore*0.9925 {
			continue
		}
		distance := angle
		if score > zeroScore {
			distance += 0.1
		}
		if distance > bestDistance || (distance == bestDistance && score > bestScore) {
			bestAngle = sign * angle
			bestDistance = distance
			bestScore = score
		}
	}
	if bestDistance > 0 {
		return bestAngle, true
	}
	return 0, false
}

func lineClusterStrong(cluster angleCluster, cols int) bool {
	if cluster.inliers >= 5 && cluster.dispersion <= 0.15 && cluster.inlierWeight >= float64(cols)*2 {
		return true
	}
	return cluster.inlierWeight >= float64(cols)*6
}

func weakLineAngleEvidence(lineAngle, lineDispersion, bestScore, zeroScore, scoreSpan float64) bool {
	if zeroScore <= 0 || bestScore <= 0 {
		return false
	}
	if math.Abs(lineAngle) > 2.5 {
		return false
	}
	if lineDispersion > 0.2 {
		return false
	}
	improvement := (bestScore - zeroScore) / zeroScore
	spanRatio := scoreSpan / zeroScore
	return improvement <= 0.12 && spanRatio <= 0.05
}

func clusterAgreementCount(target angleCluster, clusters ...angleCluster) int {
	count := 0
	for _, cluster := range clusters {
		if !cluster.ok {
			continue
		}
		if math.Abs(cluster.median-target.median) <= 0.75 {
			count++
		}
	}
	return count
}

type weightedAngleSample struct {
	angle  float64
	weight float64
}

type angleCluster struct {
	source       string
	peak         float64
	median       float64
	dispersion   float64
	totalWeight  float64
	inlierWeight float64
	count        int
	inliers      int
	ok           bool
}

type lineAngleEvidence struct {
	angle      float64
	dispersion float64
	ok         bool
	houghOnly  bool
	source     string
}

func estimateSkewTextLines(invBinary gocv.Mat, inPath string) (lineAngleEvidence, error) {
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	if rows == 0 || cols == 0 {
		return lineAngleEvidence{}, nil
	}

	componentSamples, componentWeight, err := estimateSkewComponentLines(invBinary)
	if err != nil {
		return lineAngleEvidence{}, fmt.Errorf("estimate component lines: %w", err)
	}

	baseKernelW := cols / 25
	if baseKernelW < 15 {
		baseKernelW = 15
	}
	if baseKernelW > 63 {
		baseKernelW = 63
	}

	contourSamples, houghSamples, houghWeight, err := estimateSkewMergedLineSamples(invBinary, rows, cols, baseKernelW)
	if err != nil {
		return lineAngleEvidence{}, err
	}

	totalWeight := sampleWeight(contourSamples) + componentWeight + houghWeight
	totalSamples := len(contourSamples) + len(componentSamples) + len(houghSamples)

	fallbackKernelW := cols / 18
	if fallbackKernelW <= baseKernelW {
		fallbackKernelW = baseKernelW + 8
	}
	if fallbackKernelW > 95 {
		fallbackKernelW = 95
	}
	if (len(contourSamples) < 2 || len(houghSamples) < 2 || totalSamples < 12) && fallbackKernelW > baseKernelW {
		fallbackContours, fallbackHough, fallbackHoughWeight, err := estimateSkewMergedLineSamples(invBinary, rows, cols, fallbackKernelW)
		if err != nil {
			return lineAngleEvidence{}, err
		}
		if len(fallbackContours)+len(fallbackHough) > 0 {
			deskewLogf("[%s] line-angle retry with wider merge kernel=%d", inPath, fallbackKernelW)
		}
		contourSamples = append(contourSamples, fallbackContours...)
		houghSamples = append(houghSamples, fallbackHough...)
		houghWeight += fallbackHoughWeight
		totalWeight = sampleWeight(contourSamples) + componentWeight + houghWeight
		totalSamples = len(contourSamples) + len(componentSamples) + len(houghSamples)
	}

	if totalSamples < 3 || (totalSamples < 6 && totalWeight < float64(cols)*0.75) {
		deskewLogf("[%s] line-angle unavailable samples=%d totalWeight=%.1f", inPath, totalSamples, totalWeight)
		return lineAngleEvidence{}, nil
	}

	if len(contourSamples) < 2 {
		contourSamples = nil
	}

	contourCluster := summarizeAngleCluster("contours", contourSamples, deskewAngleStep, 0.75)
	if contourCluster.inliers < 2 {
		contourSamples = nil
		contourCluster = summarizeAngleCluster("contours", contourSamples, deskewAngleStep, 0.75)
	}

	componentCluster := summarizeAngleCluster("components", componentSamples, deskewAngleStep, 0.75)
	houghCluster := summarizeAngleCluster("hough", houghSamples, deskewAngleStep, 0.75)
	combinedSamples := append([]weightedAngleSample(nil), contourSamples...)
	combinedSamples = append(combinedSamples, componentSamples...)
	combinedSamples = append(combinedSamples, houghSamples...)
	combinedCluster := summarizeAngleCluster("combined", combinedSamples, deskewAngleStep, 0.75)

	logAngleCluster(inPath, contourCluster)
	logAngleCluster(inPath, componentCluster)
	logAngleCluster(inPath, houghCluster)
	logAngleCluster(inPath, combinedCluster)

	best := chooseAngleCluster(contourCluster, componentCluster, houghCluster, combinedCluster)
	if best.source == "combined" && clusterAgreementCount(best, contourCluster, componentCluster, houghCluster) == 0 {
		deskewLogf("[%s] line-angle rejecting inconsistent combined cluster median=%.3f", inPath, best.median)
		best = chooseAngleCluster(contourCluster, componentCluster, houghCluster)
	}
	if (best.source == "combined" || best.source == "contours") &&
		math.Abs(best.median) > 1.0 &&
		math.Abs(best.median) <= 1.25 &&
		contourCluster.inliers > 0 &&
		contourCluster.inliers <= 2 &&
		componentCluster.inliers <= 1 &&
		houghCluster.inlierWeight > 0 &&
		contourCluster.inlierWeight >= houghCluster.inlierWeight*10 {
		deskewLogf("[%s] line-angle rejecting contour-dominant small angle median=%.3f", inPath, best.median)
		return lineAngleEvidence{}, nil
	}
	if best.source == "hough" && math.Abs(best.median) >= 2.5 && contourCluster.inliers < 2 && componentCluster.inliers < 2 {
		deskewLogf("[%s] line-angle rejecting isolated hough-only large angle median=%.3f", inPath, best.median)
		return lineAngleEvidence{}, nil
	}
	if !best.ok || best.inliers < 3 || !lineClusterStrong(best, cols) {
		deskewLogf("[%s] line-angle rejected weak cluster source=%s peak=%.3f inliers=%d inlierWeight=%.1f", inPath, best.source, best.peak, best.inliers, best.inlierWeight)
		return lineAngleEvidence{}, nil
	}
	deskewLogf("[%s] line-angle chose source=%s median=%.3f dispersion=%.3f", inPath, best.source, best.median, best.dispersion)

	if best.dispersion > 1.25 {
		deskewLogf("[%s] line-angle rejected due to inlier dispersion %.3f", inPath, best.dispersion)
		return lineAngleEvidence{dispersion: best.dispersion}, nil
	}

	angle := best.median
	if best.source == "combined" &&
		contourCluster.inliers == 0 &&
		componentCluster.inliers <= 1 &&
		houghCluster.inliers >= 3 &&
		math.Abs(angle) >= 0.75 &&
		math.Abs(angle) < 1.5 {
		factor := 0.75
		if math.Abs(angle) < 1.0 {
			factor = 0.25
		}
		angle *= factor
		deskewLogf("[%s] line-angle damping sparse mixed small angle to %.3f", inPath, angle)
	}

	return lineAngleEvidence{
		angle:      angle,
		dispersion: best.dispersion,
		ok:         true,
		houghOnly:  best.source == "hough",
		source:     best.source,
	}, nil
}

func dampHoughOnlyAngle(lineAngle, lineDispersion, refinedAngle, scoreSpan, zeroScore float64, inPath string, lineOK bool, houghOnly bool, textMask gocv.Mat) (float64, bool) {
	if !lineOK {
		return 0, false
	}
	if !houghOnly {
		return 0, false
	}
	if math.Abs(lineAngle) < 2.0 || lineDispersion < 0.2 {
		return 0, false
	}
	if math.Abs(refinedAngle-lineAngle) > 1.0 {
		return 0, false
	}
	if zeroScore <= 0 || scoreSpan/zeroScore < 0.25 {
		return 0, false
	}

	refineWeight := 0.7
	if lineDispersion >= 0.25 {
		refineWeight = 0.55
	}
	if math.Abs(refinedAngle) > math.Abs(lineAngle) {
		damped := lineAngle + refineWeight*(refinedAngle-lineAngle)
		deskewLogf("[%s] line-angle damped hough-only refine line=%.3f refined=%.3f dispersion=%.3f damped=%.3f", inPath, lineAngle, refinedAngle, lineDispersion, damped)
		return damped, true
	}
	damped := refinedAngle + refineWeight*(lineAngle-refinedAngle)
	deskewLogf("[%s] line-angle damped hough-only estimate line=%.3f refined=%.3f dispersion=%.3f damped=%.3f", inPath, lineAngle, refinedAngle, lineDispersion, damped)
	return damped, true
}

func estimateSparseRuledSkew(invBinary gocv.Mat, inPath string) (float64, bool, error) {
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	if rows == 0 || cols == 0 {
		return 0, false, nil
	}

	hKernelW := cols / 14
	minHKernelW := cols / 40
	if minHKernelW < 1 {
		minHKernelW = 1
	}
	if hKernelW < minHKernelW {
		hKernelW = minHKernelW
	}
	if hKernelW%2 == 0 {
		hKernelW++
	}
	maxHKernelW := cols / 6
	if maxHKernelW < 1 {
		maxHKernelW = 1
	}
	if hKernelW > maxHKernelW {
		hKernelW = maxHKernelW
	}
	vKernelH := rows / 14
	minVKernelH := rows / 40
	if minVKernelH < 1 {
		minVKernelH = 1
	}
	if vKernelH < minVKernelH {
		vKernelH = minVKernelH
	}
	if vKernelH%2 == 0 {
		vKernelH++
	}
	maxVKernelH := rows / 6
	if maxVKernelH < 1 {
		maxVKernelH = 1
	}
	if vKernelH > maxVKernelH {
		vKernelH = maxVKernelH
	}

	hKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(hKernelW, 3))
	defer hKernel.Close()
	vKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, vKernelH))
	defer vKernel.Close()

	horizontal := gocv.NewMat()
	defer horizontal.Close()
	if err := gocv.MorphologyEx(invBinary, &horizontal, gocv.MorphOpen, hKernel); err != nil {
		return 0, false, fmt.Errorf("extract horizontal ruled lines: %w", err)
	}
	vertical := gocv.NewMat()
	defer vertical.Close()
	if err := gocv.MorphologyEx(invBinary, &vertical, gocv.MorphOpen, vKernel); err != nil {
		return 0, false, fmt.Errorf("extract vertical ruled lines: %w", err)
	}
	hCoverage, hSpread := ruledMaskCoverage(horizontal, false)
	vCoverage, vSpread := ruledMaskCoverage(vertical, true)
	deskewLogf("[%s] ruled-angle coverage horizontal=%.3f spread=%.3f vertical=%.3f spread=%.3f", inPath, hCoverage, hSpread, vCoverage, vSpread)
	if hCoverage < 0.015 || vCoverage < 0.01 {
		return 0, false, nil
	}
	if hSpread < 0.03 || vSpread < 0.02 {
		return 0, false, nil
	}

	hSamples, hWeight, err := estimateSparseRuledLineSamples(horizontal, false)
	if err != nil {
		return 0, false, err
	}
	vSamples, vWeight, err := estimateSparseRuledLineSamples(vertical, true)
	if err != nil {
		return 0, false, err
	}
	hCluster := summarizeAngleCluster("ruled-horizontal", hSamples, deskewAngleStep, 0.5)
	vCluster := summarizeAngleCluster("ruled-vertical", vSamples, deskewAngleStep, 0.5)
	deskewLogf("[%s] ruled-angle horizontal samples=%d totalWeight=%.1f peak=%.3f inliers=%d inlierWeight=%.1f median=%.3f dispersion=%.3f", inPath, hCluster.count, hWeight, hCluster.peak, hCluster.inliers, hCluster.inlierWeight, hCluster.median, hCluster.dispersion)
	deskewLogf("[%s] ruled-angle vertical samples=%d totalWeight=%.1f peak=%.3f inliers=%d inlierWeight=%.1f median=%.3f dispersion=%.3f", inPath, vCluster.count, vWeight, vCluster.peak, vCluster.inliers, vCluster.inlierWeight, vCluster.median, vCluster.dispersion)
	if !hCluster.ok || !vCluster.ok {
		return 0, false, nil
	}
	if hCluster.inliers < 3 || vCluster.inliers < 3 {
		return 0, false, nil
	}
	if hCluster.dispersion > 0.35 || vCluster.dispersion > 0.35 {
		return 0, false, nil
	}
	if hCluster.inlierWeight < float64(cols)*0.35 || vCluster.inlierWeight < float64(rows)*0.35 {
		return 0, false, nil
	}
	if math.Abs(hCluster.median-vCluster.median) > 0.35 {
		deskewLogf("[%s] ruled-angle rejected inconsistent axes horizontal=%.3f vertical=%.3f", inPath, hCluster.median, vCluster.median)
		return 0, false, nil
	}
	angle := 0.5 * (hCluster.median + vCluster.median)
	dispersion := math.Max(hCluster.dispersion, vCluster.dispersion)
	deskewLogf("[%s] ruled-angle angle=%.3f dispersion=%.3f", inPath, angle, dispersion)
	return angle, true, nil
}

func ruledMaskCoverage(mask gocv.Mat, vertical bool) (float64, float64) {
	rows := mask.Rows()
	cols := mask.Cols()
	if rows == 0 || cols == 0 {
		return 0, 0
	}

	nz := gocv.CountNonZero(mask)
	coverage := float64(nz) / float64(rows*cols)

	if !vertical {
		activeRows := 0
		minInk := cols / 80
		if minInk < 1 {
			minInk = 1
		}
		for y := 0; y < rows; y++ {
			row := mask.RowRange(y, y+1)
			rowInk := gocv.CountNonZero(row)
			row.Close()
			if rowInk >= minInk {
				activeRows++
			}
		}
		return coverage, float64(activeRows) / float64(rows)
	}

	activeCols := 0
	minInk := rows / 80
	if minInk < 1 {
		minInk = 1
	}
	for x := 0; x < cols; x++ {
		col := mask.ColRange(x, x+1)
		colInk := gocv.CountNonZero(col)
		col.Close()
		if colInk >= minInk {
			activeCols++
		}
	}
	return coverage, float64(activeCols) / float64(cols)
}

func estimateSparseRuledLineSamples(mask gocv.Mat, verticalMode bool) ([]weightedAngleSample, float64, error) {
	rows := mask.Rows()
	cols := mask.Cols()
	if rows == 0 || cols == 0 {
		return nil, 0, nil
	}

	samples := make([]weightedAngleSample, 0, 32)
	totalWeight := 0.0

	contours := gocv.FindContours(mask, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	minSpan := int(math.Round(float64(cols) * 0.20))
	maxThickness := int(math.Round(float64(rows) * 0.04))
	if verticalMode {
		minSpan = int(math.Round(float64(rows) * 0.20))
		maxThickness = int(math.Round(float64(cols) * 0.04))
	}
	if maxThickness < 1 {
		maxThickness = 1
	}
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		r := gocv.BoundingRect(c)
		if !verticalMode {
			if r.Dx() < minSpan || r.Dy() < 1 || r.Dy() > maxThickness {
				continue
			}
		} else {
			if r.Dy() < minSpan || r.Dx() < 1 || r.Dx() > maxThickness {
				continue
			}
		}

		rect := gocv.MinAreaRect(c)
		angle := normalizeRectAngle(rect.Angle, rect.Width, rect.Height)
		if verticalMode {
			if angle < 0 {
				angle += 90
			} else {
				angle -= 90
			}
		}
		for angle <= -45 {
			angle += 90
		}
		for angle > 45 {
			angle -= 90
		}
		if math.Abs(angle) > deskewProjectionLimit {
			continue
		}

		span := r.Dx()
		thickness := r.Dy()
		if verticalMode {
			span = r.Dy()
			thickness = r.Dx()
		}
		weight := float64(span) * math.Max(1, math.Sqrt(float64(thickness)))
		samples = append(samples, weightedAngleSample{angle: angle, weight: weight})
		totalWeight += weight
	}

	lines := gocv.NewMat()
	defer lines.Close()

	threshold := cols / 12
	minLineLength := float32(cols) * 0.20
	maxLineGap := float32(cols) * 0.03
	if verticalMode {
		threshold = rows / 12
		minLineLength = float32(rows) * 0.20
		maxLineGap = float32(rows) * 0.03
	}
	if threshold < 1 {
		threshold = 1
	}
	if err := gocv.HoughLinesPWithParams(mask, &lines, 1, math.Pi/180, threshold, minLineLength, maxLineGap); err != nil {
		return nil, 0, fmt.Errorf("estimate ruled hough lines: %w", err)
	}

	for i := 0; i < lines.Rows(); i++ {
		x1 := lines.GetIntAt(i, 0)
		y1 := lines.GetIntAt(i, 1)
		x2 := lines.GetIntAt(i, 2)
		y2 := lines.GetIntAt(i, 3)

		dx := float64(x2 - x1)
		dy := float64(y2 - y1)
		if dx == 0 && dy == 0 {
			continue
		}
		if !verticalMode && math.Abs(dx) < math.Abs(dy) {
			continue
		}
		if verticalMode && math.Abs(dy) < math.Abs(dx) {
			continue
		}

		angle := math.Atan2(dy, dx) * 180 / math.Pi
		for angle <= -90 {
			angle += 180
		}
		for angle > 90 {
			angle -= 180
		}
		if verticalMode {
			if angle < 0 {
				angle += 90
			} else {
				angle -= 90
			}
		}
		for angle <= -45 {
			angle += 90
		}
		for angle > 45 {
			angle -= 90
		}
		if math.Abs(angle) > deskewProjectionLimit {
			continue
		}

		length := math.Hypot(dx, dy)
		if length < float64(minLineLength) {
			continue
		}

		weight := math.Min(length, math.Max(float64(cols), float64(rows))*0.6)
		samples = append(samples, weightedAngleSample{angle: angle, weight: weight})
		totalWeight += weight
	}

	return samples, totalWeight, nil
}

type deskewComponent struct {
	cx   float64
	cy   float64
	w    float64
	h    float64
	area float64
}

func buildDeskewTextMask(invBinary gocv.Mat) (gocv.Mat, bool, error) {
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	if rows == 0 || cols == 0 {
		return gocv.Mat{}, false, nil
	}

	contours := gocv.FindContours(invBinary, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	if contours.Size() == 0 {
		return gocv.Mat{}, false, nil
	}

	mask := gocv.Zeros(rows, cols, gocv.MatTypeCV8U)
	kept := 0
	minArea := math.Max(3, float64(rows*cols)*0.000002)
	maxArea := float64(rows*cols) * 0.0025
	maxWidth := int(math.Round(float64(cols) * 0.12))
	maxHeight := int(math.Round(float64(rows) * 0.05))

	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		area := gocv.ContourArea(c)
		if area < minArea || area > maxArea {
			continue
		}

		r := gocv.BoundingRect(c)
		w := r.Dx()
		h := r.Dy()
		if w < 2 || h < 2 || w > maxWidth || h > maxHeight {
			continue
		}
		if r.Min.X <= 1 || r.Min.Y <= 1 || r.Max.X >= cols-1 || r.Max.Y >= rows-1 {
			continue
		}

		aspect := float64(w) / float64(h)
		if aspect > 12 || aspect < 0.08 {
			continue
		}

		rectArea := float64(w * h)
		if rectArea <= 0 {
			continue
		}
		fill := area / rectArea
		if fill > 0.9 {
			continue
		}

		if err := gocv.DrawContours(&mask, contours, i, color.RGBA{R: 255, G: 255, B: 255, A: 255}, -1); err != nil {
			mask.Close()
			return gocv.Mat{}, false, fmt.Errorf("draw deskew text contour %d: %w", i, err)
		}
		kept++
	}

	if kept < 32 {
		mask.Close()
		return gocv.Mat{}, false, nil
	}

	return mask, true, nil
}

func estimateSkewMergedLineSamples(invBinary gocv.Mat, rows, cols, kernelW int) ([]weightedAngleSample, []weightedAngleSample, float64, error) {
	if kernelW < 3 {
		kernelW = 3
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelW, 3))
	defer kernel.Close()

	merged := gocv.NewMat()
	defer merged.Close()
	if err := gocv.MorphologyEx(invBinary, &merged, gocv.MorphClose, kernel); err != nil {
		return nil, nil, 0, fmt.Errorf("merge text lines: %w", err)
	}

	contours := gocv.FindContours(merged, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	minWidth := int(math.Round(float64(cols) * 0.18))
	minArea := float64(rows*cols) * 0.00025
	contourSamples := make([]weightedAngleSample, 0, contours.Size())

	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		r := gocv.BoundingRect(c)
		if r.Dx() < minWidth || r.Dy() < 6 || r.Dy() > rows/5 {
			continue
		}

		area := gocv.ContourArea(c)
		if area < minArea {
			continue
		}

		rect := gocv.MinAreaRect(c)
		observed := normalizeRectAngle(rect.Angle, rect.Width, rect.Height)
		if math.Abs(observed) > deskewLineAngleLimit {
			continue
		}

		weight := math.Sqrt(area) * float64(r.Dx())
		contourSamples = append(contourSamples, weightedAngleSample{
			angle:  observed,
			weight: weight,
		})
	}

	houghSamples, houghWeight, err := estimateSkewHoughLines(merged)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("estimate hough lines: %w", err)
	}

	return contourSamples, houghSamples, houghWeight, nil
}

func estimateSkewComponentLines(invBinary gocv.Mat) ([]weightedAngleSample, float64, error) {
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	if rows == 0 || cols == 0 {
		return nil, 0, nil
	}

	components := collectDeskewComponents(invBinary)
	if len(components) < 24 {
		return nil, 0, nil
	}

	widths := make([]float64, 0, len(components))
	heights := make([]float64, 0, len(components))
	for _, c := range components {
		widths = append(widths, c.w)
		heights = append(heights, c.h)
	}
	medianW := medianFloat64(widths)
	medianH := medianFloat64(heights)
	if medianW <= 0 || medianH <= 0 {
		return nil, 0, nil
	}

	sort.Slice(components, func(i, j int) bool {
		if components[i].cx == components[j].cx {
			return components[i].cy < components[j].cy
		}
		return components[i].cx < components[j].cx
	})

	parent := make([]int, len(components))
	for i := range parent {
		parent[i] = i
	}
	find := func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	union := func(a, b int) {
		ra := find(a)
		rb := find(b)
		if ra != rb {
			parent[rb] = ra
		}
	}

	maxGap := math.Max(medianW*8, float64(cols)*0.025)
	maxDeltaY := math.Max(medianH*1.6, 6)
	maxSlope := math.Tan(1.75 * math.Pi / 180)
	for i := 0; i < len(components); i++ {
		for j := i + 1; j < len(components); j++ {
			dx := components[j].cx - components[i].cx
			if dx > maxGap {
				break
			}
			dy := math.Abs(components[j].cy - components[i].cy)
			allowedDeltaY := maxDeltaY + dx*maxSlope
			if dy > allowedDeltaY {
				continue
			}
			union(i, j)
		}
	}

	groups := make(map[int][]deskewComponent)
	for i, c := range components {
		root := find(i)
		groups[root] = append(groups[root], c)
	}

	samples := make([]weightedAngleSample, 0, len(groups))
	totalWeight := 0.0
	for _, group := range groups {
		if len(group) < 6 {
			continue
		}

		minX := group[0].cx
		maxX := group[0].cx
		for _, c := range group[1:] {
			if c.cx < minX {
				minX = c.cx
			}
			if c.cx > maxX {
				maxX = c.cx
			}
		}
		span := maxX - minX
		if span < float64(cols)*0.18 {
			continue
		}

		angle, residual, ok := fitDeskewLine(group)
		if !ok || math.Abs(angle) > deskewLineAngleLimit {
			continue
		}
		if residual > math.Max(medianH*0.9, 4) {
			continue
		}

		weight := span * float64(len(group))
		samples = append(samples, weightedAngleSample{
			angle:  angle,
			weight: weight,
		})
		totalWeight += weight
	}

	return samples, totalWeight, nil
}

func collectDeskewComponents(invBinary gocv.Mat) []deskewComponent {
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	contours := gocv.FindContours(invBinary, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	maxArea := float64(rows*cols) * 0.0025
	maxWidth := float64(cols) * 0.08
	maxHeight := float64(rows) * 0.04
	components := make([]deskewComponent, 0, contours.Size())
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		area := gocv.ContourArea(c)
		if area < 2 || area > maxArea {
			continue
		}

		r := gocv.BoundingRect(c)
		w := float64(r.Dx())
		h := float64(r.Dy())
		if w < 2 || h < 2 || w > maxWidth || h > maxHeight {
			continue
		}

		aspect := w / h
		if aspect < 0.08 || aspect > 8 {
			continue
		}

		components = append(components, deskewComponent{
			cx:   float64(r.Min.X+r.Max.X) / 2,
			cy:   float64(r.Min.Y+r.Max.Y) / 2,
			w:    w,
			h:    h,
			area: area,
		})
	}
	return components
}

func fitDeskewLine(group []deskewComponent) (float64, float64, bool) {
	if len(group) < 2 {
		return 0, 0, false
	}

	sumW := 0.0
	sumX := 0.0
	sumY := 0.0
	for _, c := range group {
		w := math.Max(c.area, 1)
		sumW += w
		sumX += c.cx * w
		sumY += c.cy * w
	}
	if sumW == 0 {
		return 0, 0, false
	}

	meanX := sumX / sumW
	meanY := sumY / sumW
	num := 0.0
	den := 0.0
	for _, c := range group {
		w := math.Max(c.area, 1)
		dx := c.cx - meanX
		dy := c.cy - meanY
		num += w * dx * dy
		den += w * dx * dx
	}
	if den == 0 {
		return 0, 0, false
	}

	slope := num / den
	angle := math.Atan(slope) * 180 / math.Pi
	residuals := make([]float64, 0, len(group))
	for _, c := range group {
		pred := meanY + slope*(c.cx-meanX)
		residuals = append(residuals, math.Abs(c.cy-pred))
	}
	return angle, medianFloat64(residuals), true
}

func medianFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	mid := len(ordered) / 2
	if len(ordered)%2 == 1 {
		return ordered[mid]
	}
	return 0.5 * (ordered[mid-1] + ordered[mid])
}

func estimateSkewHoughLines(merged gocv.Mat) ([]weightedAngleSample, float64, error) {
	rows := merged.Rows()
	cols := merged.Cols()
	if rows == 0 || cols == 0 {
		return nil, 0, nil
	}

	lines := gocv.NewMat()
	defer lines.Close()

	threshold := cols / 10
	if threshold < 20 {
		threshold = 20
	}

	minLineLength := float32(cols) * 0.12
	maxLineGap := float32(cols) * 0.04
	if err := gocv.HoughLinesPWithParams(merged, &lines, 1, math.Pi/180, threshold, minLineLength, maxLineGap); err != nil {
		return nil, 0, err
	}
	if lines.Empty() {
		return nil, 0, nil
	}

	samples := make([]weightedAngleSample, 0, lines.Rows())
	totalWeight := 0.0
	maxAngle := deskewLineAngleLimit

	for i := 0; i < lines.Rows(); i++ {
		x1 := lines.GetIntAt(i, 0)
		y1 := lines.GetIntAt(i, 1)
		x2 := lines.GetIntAt(i, 2)
		y2 := lines.GetIntAt(i, 3)

		dx := float64(x2 - x1)
		dy := float64(y2 - y1)
		if dx == 0 && dy == 0 {
			continue
		}

		angle := math.Atan2(dy, dx) * 180 / math.Pi
		for angle <= -90 {
			angle += 180
		}
		for angle > 90 {
			angle -= 180
		}
		if angle < -45 {
			angle += 90
		}
		if angle > 45 {
			angle -= 90
		}
		if math.Abs(angle) > maxAngle {
			continue
		}

		length := math.Hypot(dx, dy)
		if length < float64(minLineLength) {
			continue
		}

		tilt := math.Abs(angle)
		if tilt < 0.1 {
			continue
		}

		tiltWeight := tilt
		if tiltWeight < 0.2 {
			tiltWeight = 0.2
		}
		effectiveLength := length
		maxWeightedLength := float64(cols) * 0.35
		if effectiveLength > maxWeightedLength {
			effectiveLength = maxWeightedLength
		}
		weight := effectiveLength * tiltWeight

		samples = append(samples, weightedAngleSample{
			angle:  angle,
			weight: weight,
		})
		totalWeight += weight
	}

	return samples, totalWeight, nil
}

func normalizeRectAngle(angle float64, width, height int) float64 {
	if width < height {
		angle += 90
	}
	for angle <= -45 {
		angle += 90
	}
	for angle > 45 {
		angle -= 90
	}
	return angle
}

func weightedMedianAngle(samples []weightedAngleSample) float64 {
	ordered := append([]weightedAngleSample(nil), samples...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].angle < ordered[j].angle
	})

	total := 0.0
	for _, sample := range ordered {
		total += sample.weight
	}

	threshold := total / 2
	acc := 0.0
	for _, sample := range ordered {
		acc += sample.weight
		if acc >= threshold {
			return sample.angle
		}
	}

	return ordered[len(ordered)-1].angle
}

func weightedMedianAbsDeviation(samples []weightedAngleSample, center float64) float64 {
	deviations := make([]weightedAngleSample, 0, len(samples))
	for _, sample := range samples {
		deviations = append(deviations, weightedAngleSample{
			angle:  math.Abs(sample.angle - center),
			weight: sample.weight,
		})
	}
	return weightedMedianAngle(deviations)
}

func sampleWeight(samples []weightedAngleSample) float64 {
	total := 0.0
	for _, sample := range samples {
		total += sample.weight
	}
	return total
}

func dominantAngleCluster(samples []weightedAngleSample, step float64, radius float64) (float64, []weightedAngleSample, float64) {
	if len(samples) == 0 {
		return 0, nil, 0
	}

	if step <= 0 {
		step = 0.25
	}
	if radius <= 0 {
		radius = step * 2
	}

	minAngle := samples[0].angle
	maxAngle := samples[0].angle
	for _, sample := range samples[1:] {
		if sample.angle < minAngle {
			minAngle = sample.angle
		}
		if sample.angle > maxAngle {
			maxAngle = sample.angle
		}
	}

	binCount := int(math.Ceil((maxAngle-minAngle)/step)) + 1
	if binCount < 1 {
		binCount = 1
	}
	bins := make([]float64, binCount)
	for _, sample := range samples {
		idx := int(math.Round((sample.angle - minAngle) / step))
		if idx < 0 {
			idx = 0
		}
		if idx >= binCount {
			idx = binCount - 1
		}
		bins[idx] += sample.weight
	}

	bestIdx := 0
	bestScore := -1.0
	for i := range bins {
		score := bins[i]
		if i > 0 {
			score += 0.5 * bins[i-1]
		}
		if i+1 < len(bins) {
			score += 0.5 * bins[i+1]
		}
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	peak := minAngle + float64(bestIdx)*step
	inliers := make([]weightedAngleSample, 0, len(samples))
	inlierWeight := 0.0
	for _, sample := range samples {
		if math.Abs(sample.angle-peak) > radius {
			continue
		}
		inliers = append(inliers, sample)
		inlierWeight += sample.weight
	}

	return peak, inliers, inlierWeight
}

func summarizeAngleCluster(source string, samples []weightedAngleSample, step float64, radius float64) angleCluster {
	cluster := angleCluster{
		source:      source,
		totalWeight: sampleWeight(samples),
		count:       len(samples),
	}
	if len(samples) == 0 {
		return cluster
	}

	peak, inliers, inlierWeight := dominantAngleCluster(samples, step, radius)
	cluster.peak = peak
	cluster.inliers = len(inliers)
	cluster.inlierWeight = inlierWeight
	if len(inliers) == 0 {
		return cluster
	}

	cluster.median = weightedMedianAngle(inliers)
	cluster.dispersion = weightedMedianAbsDeviation(inliers, cluster.median)
	cluster.ok = true
	return cluster
}

func logAngleCluster(inPath string, cluster angleCluster) {
	deskewLogf("[%s] line-angle %s samples=%d totalWeight=%.1f peak=%.3f inliers=%d inlierWeight=%.1f median=%.3f dispersion=%.3f", inPath, cluster.source, cluster.count, cluster.totalWeight, cluster.peak, cluster.inliers, cluster.inlierWeight, cluster.median, cluster.dispersion)
}

func chooseAngleCluster(clusters ...angleCluster) angleCluster {
	best := angleCluster{}
	bestScore := -1.0
	type scoredCluster struct {
		cluster angleCluster
		score   float64
	}
	scored := make([]scoredCluster, 0, len(clusters))
	for _, cluster := range clusters {
		if !cluster.ok {
			continue
		}

		score := cluster.inlierWeight / (1 + cluster.dispersion)
		scored = append(scored, scoredCluster{cluster: cluster, score: score})
		if score > bestScore {
			best = cluster
			bestScore = score
		}
	}
	if best.source == "hough" {
		for _, candidate := range scored {
			cluster := candidate.cluster
			if cluster.source == "hough" {
				continue
			}
			if candidate.score < bestScore*0.85 {
				continue
			}
			if math.Abs(cluster.median-best.median) < 1.0 {
				continue
			}
			if cluster.dispersion > best.dispersion {
				continue
			}
			best = cluster
			bestScore = candidate.score
		}
	}
	return best
}

func adaptiveBinary(gray gocv.Mat) (gocv.Mat, error) {
	blur := gocv.NewMat()
	if err := gocv.GaussianBlur(gray, &blur, image.Pt(3, 3), 0, 0, gocv.BorderDefault); err != nil {
		return blur, fmt.Errorf("gaussian blur: %w", err)
	}

	bin := gocv.NewMat()
	// THRESH_BINARY: black text becomes 0, background 255 (usually), then we invert later
	if err := gocv.AdaptiveThreshold(
		blur,
		&bin,
		255,
		gocv.AdaptiveThresholdGaussian,
		gocv.ThresholdBinary,
		41,
		15,
	); err != nil {
		return bin, fmt.Errorf("adaptive threshold: %w", err)
	}

	blur.Close()
	return bin, nil
}

func contentBBox(binary gocv.Mat) (image.Rectangle, bool) {
	inv := gocv.NewMat()
	defer inv.Close()
	if err := gocv.BitwiseNot(binary, &inv); err != nil {
		return image.Rectangle{}, false
	}

	k := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer k.Close()

	clean := gocv.NewMat()
	defer clean.Close()
	if err := gocv.MorphologyEx(inv, &clean, gocv.MorphOpen, k); err != nil {
		return image.Rectangle{}, false
	}

	contours := gocv.FindContours(clean, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	if contours.Size() == 0 {
		return image.Rectangle{}, false
	}

	h, w := binary.Rows(), binary.Cols()
	minArea := float64(h*w) * 0.001

	xs, ys := w, h
	xe, ye := 0, 0
	kept := 0

	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		area := gocv.ContourArea(c)
		if area < minArea {
			continue
		}
		r := gocv.BoundingRect(c)
		if r.Min.X < xs {
			xs = r.Min.X
		}
		if r.Min.Y < ys {
			ys = r.Min.Y
		}
		if r.Max.X > xe {
			xe = r.Max.X
		}
		if r.Max.Y > ye {
			ye = r.Max.Y
		}
		kept++
	}

	if kept == 0 {
		return image.Rectangle{}, false
	}

	pad := int(math.Round(0.01 * float64(lo.Ternary(h > w, h, w))))
	xs = lo.Ternary(xs-pad > 0, xs-pad, 0)
	ys = lo.Ternary(ys-pad > 0, ys-pad, 0)
	xe = lo.Ternary(xe+pad < w, xe+pad, w)
	ye = lo.Ternary(ye+pad < h, ye+pad, h)

	if xe-xs < 50 || ye-ys < 50 {
		return image.Rectangle{}, false
	}

	return image.Rect(xs, ys, xe, ye), true
}

func projectionVariance(invBinary gocv.Mat) float64 {
	rowScore := projectionVarianceRows(invBinary)
	bodyScore := projectionVarianceBodyBands(invBinary)
	columnScore := projectionVarianceColumnBands(invBinary)
	return math.Max(math.Max(rowScore, bodyScore), columnScore)
}

func projectionVarianceRows(invBinary gocv.Mat) float64 {
	// invBinary: ink pixels are >0, background 0.
	// Score = variance of row sums on text-bearing rows only.
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	if rows == 0 || cols == 0 {
		return 0
	}

	sums := make([]float64, 0, rows)
	minInk := cols / 500
	if minInk < 2 {
		minInk = 2
	}

	for y := 0; y < rows; y++ {
		row := invBinary.RowRange(y, y+1)
		nz := gocv.CountNonZero(row)
		row.Close()
		if nz < minInk {
			continue
		}
		sums = append(sums, float64(nz)/float64(cols))
	}

	if len(sums) < 8 {
		sums = sums[:0]
		for y := 0; y < rows; y++ {
			row := invBinary.RowRange(y, y+1)
			nz := gocv.CountNonZero(row)
			row.Close()
			sums = append(sums, float64(nz)/float64(cols))
		}
	}

	mean := 0.0
	for _, v := range sums {
		mean += v
	}
	mean /= float64(len(sums))

	variance := 0.0
	for _, v := range sums {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(sums))
	return variance
}

type deskewBand struct {
	start int
	end   int
}

func projectionVarianceBodyBands(invBinary gocv.Mat) float64 {
	bands := detectDeskewBodyBands(invBinary)
	if len(bands) == 0 {
		return 0
	}

	totalWeight := 0.0
	totalScore := 0.0
	for _, band := range bands {
		height := band.end - band.start
		if height < 12 {
			continue
		}

		region := invBinary.Region(image.Rect(0, band.start, invBinary.Cols(), band.end))
		score := projectionVarianceRows(region)
		region.Close()
		if score <= 0 {
			continue
		}

		weight := float64(height)
		totalWeight += weight
		totalScore += score * weight
	}

	if totalWeight == 0 {
		return 0
	}
	return totalScore / totalWeight
}

func projectionVarianceColumnBands(invBinary gocv.Mat) float64 {
	bands := detectDeskewColumnBands(invBinary)
	if len(bands) <= 1 {
		return 0
	}

	totalWeight := 0.0
	totalScore := 0.0
	for _, band := range bands {
		width := band.end - band.start
		if width < 16 {
			continue
		}

		region := invBinary.Region(image.Rect(band.start, 0, band.end, invBinary.Rows()))
		score := math.Max(projectionVarianceRows(region), projectionVarianceBodyBands(region))
		region.Close()
		if score <= 0 {
			continue
		}

		weight := float64(width)
		totalWeight += weight
		totalScore += score * weight
	}

	if totalWeight == 0 {
		return 0
	}
	return totalScore / totalWeight
}

func detectDeskewBodyBands(invBinary gocv.Mat) []deskewBand {
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	if rows == 0 || cols == 0 {
		return nil
	}

	rowInk := make([]int, rows)
	maxInk := 0
	for y := 0; y < rows; y++ {
		row := invBinary.RowRange(y, y+1)
		nz := gocv.CountNonZero(row)
		row.Close()
		rowInk[y] = nz
		if nz > maxInk {
			maxInk = nz
		}
	}
	if maxInk == 0 {
		return nil
	}

	activeThreshold := cols / 40
	if activeThreshold < 8 {
		activeThreshold = 8
	}
	strongThreshold := cols / 18
	if strongThreshold < activeThreshold {
		strongThreshold = activeThreshold
	}

	var bands []deskewBand
	start := -1
	strongRows := 0
	gap := 0
	maxGap := 6
	minHeight := rows / 18
	if minHeight < 24 {
		minHeight = 24
	}
	for y, ink := range rowInk {
		active := ink >= activeThreshold
		if active {
			if start < 0 {
				start = y
				strongRows = 0
				gap = 0
			}
			gap = 0
			if ink >= strongThreshold {
				strongRows++
			}
			continue
		}
		if start < 0 {
			continue
		}
		gap++
		if gap <= maxGap {
			continue
		}
		end := y - gap + 1
		if end-start >= minHeight && strongRows >= 6 {
			bands = append(bands, deskewBand{start: start, end: end})
		}
		start = -1
		strongRows = 0
		gap = 0
	}
	if start >= 0 {
		end := rows
		if end-start >= minHeight && strongRows >= 6 {
			bands = append(bands, deskewBand{start: start, end: end})
		}
	}

	if len(bands) <= 1 {
		return bands
	}

	filtered := bands[:0]
	maxHeight := 0
	for _, band := range bands {
		height := band.end - band.start
		if height > maxHeight {
			maxHeight = height
		}
	}
	for _, band := range bands {
		height := band.end - band.start
		if height*3 < maxHeight {
			continue
		}
		filtered = append(filtered, band)
	}
	return filtered
}

func detectDeskewColumnBands(invBinary gocv.Mat) []deskewBand {
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	if rows == 0 || cols == 0 {
		return nil
	}

	colInk := make([]int, cols)
	maxInk := 0
	for x := 0; x < cols; x++ {
		col := invBinary.ColRange(x, x+1)
		nz := gocv.CountNonZero(col)
		col.Close()
		colInk[x] = nz
		if nz > maxInk {
			maxInk = nz
		}
	}
	if maxInk == 0 {
		return nil
	}

	activeThreshold := rows / 100
	if activeThreshold < 4 {
		activeThreshold = 4
	}
	strongThreshold := rows / 35
	if strongThreshold < activeThreshold {
		strongThreshold = activeThreshold
	}

	var bands []deskewBand
	start := -1
	strongCols := 0
	gap := 0
	maxGap := cols / 80
	if maxGap < 8 {
		maxGap = 8
	}
	minWidth := cols / 12
	if minWidth < 32 {
		minWidth = 32
	}
	for x, ink := range colInk {
		active := ink >= activeThreshold
		if active {
			if start < 0 {
				start = x
				strongCols = 0
				gap = 0
			}
			gap = 0
			if ink >= strongThreshold {
				strongCols++
			}
			continue
		}
		if start < 0 {
			continue
		}
		gap++
		if gap <= maxGap {
			continue
		}
		end := x - gap + 1
		if end-start >= minWidth && strongCols >= minWidth/4 {
			bands = append(bands, deskewBand{start: start, end: end})
		}
		start = -1
		strongCols = 0
		gap = 0
	}
	if start >= 0 {
		end := cols
		if end-start >= minWidth && strongCols >= minWidth/4 {
			bands = append(bands, deskewBand{start: start, end: end})
		}
	}

	if len(bands) <= 1 {
		return bands
	}

	filtered := bands[:0]
	maxWidth := 0
	for _, band := range bands {
		width := band.end - band.start
		if width > maxWidth {
			maxWidth = width
		}
	}
	for _, band := range bands {
		width := band.end - band.start
		if width*4 < maxWidth {
			continue
		}
		filtered = append(filtered, band)
	}
	return filtered
}

func rotateSameSize(src gocv.Mat, angleDeg float64) (gocv.Mat, error) {
	h, w := src.Rows(), src.Cols()
	center := image.Pt(w/2, h/2)
	M := gocv.GetRotationMatrix2D(center, angleDeg, 1.0)
	defer M.Close()

	dst := gocv.NewMat()
	if err := gocv.WarpAffineWithParams(src, &dst, M, image.Point{X: w, Y: h}, gocv.InterpolationNearestNeighbor, gocv.BorderConstant, color.RGBA{}); err != nil {
		return dst, fmt.Errorf("warp affine: %w", err)
	}
	if dst.Empty() {
		dst.Close()
		return gocv.Mat{}, errors.New("warp produced empty image")
	}
	return dst, nil
}

func rotateKeepAll(src gocv.Mat, angleDeg float64, bg uint8) (gocv.Mat, error) {
	h, w := src.Rows(), src.Cols()
	cx := float64(w) / 2.0
	cy := float64(h) / 2.0

	center := image.Pt(int(math.Round(cx)), int(math.Round(cy)))
	M := gocv.GetRotationMatrix2D(center, angleDeg, 1.0)
	defer M.Close()

	cos := math.Abs(M.GetDoubleAt(0, 0))
	sin := math.Abs(M.GetDoubleAt(0, 1))
	newW := int(math.Round(float64(h)*sin + float64(w)*cos))
	newH := int(math.Round(float64(h)*cos + float64(w)*sin))
	if newW < 1 || newH < 1 {
		return gocv.Mat{}, fmt.Errorf("rotation produced invalid size %dx%d", newW, newH)
	}

	M.SetDoubleAt(0, 2, M.GetDoubleAt(0, 2)+float64(newW)/2-cx)
	M.SetDoubleAt(1, 2, M.GetDoubleAt(1, 2)+float64(newH)/2-cy)

	dst := gocv.NewMat()
	borderVal := color.RGBA{R: bg, G: bg, B: bg, A: 0}
	if err := gocv.WarpAffineWithParams(src, &dst, M, image.Point{X: newW, Y: newH}, gocv.InterpolationCubic, gocv.BorderConstant, borderVal); err != nil {
		return dst, fmt.Errorf("warp affine: %w", err)
	}
	if dst.Empty() {
		dst.Close()
		return gocv.Mat{}, errors.New("warp produced empty image")
	}
	return dst, nil
}

func exceedsWarpAffineLimit(src gocv.Mat, angleDeg float64) bool {
	h, w := src.Rows(), src.Cols()
	if h >= opencvRemapMaxDim || w >= opencvRemapMaxDim {
		return true
	}

	rad := math.Abs(angleDeg) * math.Pi / 180.0
	cos := math.Abs(math.Cos(rad))
	sin := math.Abs(math.Sin(rad))
	newW := int(math.Round(float64(h)*sin + float64(w)*cos))
	newH := int(math.Round(float64(h)*cos + float64(w)*sin))

	return newW >= opencvRemapMaxDim || newH >= opencvRemapMaxDim
}
