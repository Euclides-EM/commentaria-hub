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
	opencvRemapMaxDim = math.MaxInt16 - 1
	deskewDebug       = false
)

func deskewLogf(format string, args ...any) {
	if !deskewDebug {
		return
	}
	log.Printf(format, args...)
}

func maxDeskewWorkers() int {
	n := runtime.NumCPU()
	if n < 2 {
		return 2
	}
	if n > 24 {
		return 24
	}
	return n
}

type projParams struct {
	DownscaleMax int     // max(w,h) for angle estimation
	TrimBorder   bool    // crop to content before estimating angle
	AngleLimit   float64 // degrees, search [-limit, +limit]
	AngleStep    float64 // degrees
	MinRotate    float64 // degrees, skip tiny rotate
	Bg           uint8   // background fill (white)
}

func defaultProjParams() projParams {
	return projParams{
		DownscaleMax: 1600,
		TrimBorder:   true,
		AngleLimit:   2.0,
		AngleStep:    0.25,
		MinRotate:    0.15,
		Bg:           255,
	}
}

func DeskewPNGs(src string, dst string) error {
	params := defaultProjParams()

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

			if err := deskewOneProjection(j.inPath, j.outPath, params); err != nil {
				return fmt.Errorf("deskew %q: %w", j.inPath, err)
			}
			return nil
		})
	}

	return grp.Wait()
}

func deskewOneProjection(inPath, outPath string, p projParams) error {
	img := gocv.IMRead(inPath, gocv.IMReadColor)
	if img.Empty() {
		return fmt.Errorf("read image %q: empty or unsupported format", inPath)
	}
	defer img.Close()

	// Work on a downscaled copy for angle estimation.
	small := img.Clone()
	defer small.Close()

	if p.DownscaleMax > 0 {
		var err error
		small, err = resizeMaxSide(small, p.DownscaleMax)
		if err != nil {
			return fmt.Errorf("downscale for angle estimation: %w", err)
		}
		// When resized, resizeMaxSide closed the clone and returned a new Mat; defer above closes current small.
	}

	angle, err := estimateSkewProjection(small, p, inPath)
	if err != nil {
		return err
	}

	deskewLogf("[%s] estimated angle=%.3f deg (minRotate=%.2f)", inPath, angle, p.MinRotate)

	if math.Abs(angle) < p.MinRotate {
		deskewLogf("[%s] skipping rotation (angle below threshold)", inPath)
		if ok := gocv.IMWrite(outPath, img); !ok {
			return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
		}
		return nil
	}

	if exceedsWarpAffineLimit(img, angle) {
		log.Printf("Skipping deskew for %q: image dimensions exceed OpenCV remap limit during rotation", inPath)
		if ok := gocv.IMWrite(outPath, img); !ok {
			return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
		}
		return nil
	}

	rot, err := rotateKeepAll(img, angle, p.Bg)
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

func estimateSkewProjection(img gocv.Mat, p projParams, inPath string) (float64, error) {
	gray := gocv.NewMat()
	defer gray.Close()
	if err := gocv.CvtColor(img, &gray, gocv.ColorBGRToGray); err != nil {
		return 0, fmt.Errorf("convert to grayscale: %w", err)
	}

	// Optional trim: crop to content bbox before scoring
	if p.TrimBorder {
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

	lineAngle, lineDispersion, lineOK, err := estimateSkewTextLines(inv, p, inPath)
	if err != nil {
		return 0, err
	}
	if lineOK {
		if lineAngle < -p.AngleLimit {
			lineAngle = -p.AngleLimit
		}
		if lineAngle > p.AngleLimit {
			lineAngle = p.AngleLimit
		}
		deskewLogf("[%s] line-angle angle=%.3f dispersion=%.3f", inPath, lineAngle, lineDispersion)
		if lineDispersion <= 0.75 {
			refinedAngle, refinedScore, zeroScore, scoreSpan, err := refineProjectionAroundAngle(inv, lineAngle, p, inPath)
			if err != nil {
				return 0, err
			}
			deskewLogf("[%s] line-angle refine candidate=%.3f score=%.6f zeroScore=%.6f scoreSpan=%.6f", inPath, refinedAngle, refinedScore, zeroScore, scoreSpan)
			if math.Abs(refinedAngle-lineAngle) <= 0.75 && refinedScore >= zeroScore*1.01 {
				return refinedAngle, nil
			}
			if scoreSpan <= zeroScore*0.01 {
				blended := 0.7*lineAngle + 0.3*refinedAngle
				if blended < -p.AngleLimit {
					blended = -p.AngleLimit
				}
				if blended > p.AngleLimit {
					blended = p.AngleLimit
				}
				deskewLogf("[%s] line-angle accepted on projection plateau blended=%.3f", inPath, blended)
				return blended, nil
			}
			if math.Abs(lineAngle) >= 0.75 && math.Abs(refinedAngle) <= 0.35 && math.Abs(refinedAngle-lineAngle) >= 0.5 && refinedScore-zeroScore <= zeroScore*0.003 {
				blended := 0.8*lineAngle + 0.2*refinedAngle
				if blended < -p.AngleLimit {
					blended = -p.AngleLimit
				}
				if blended > p.AngleLimit {
					blended = p.AngleLimit
				}
				deskewLogf("[%s] line-angle accepted on near-zero shoulder blended=%.3f", inPath, blended)
				return blended, nil
			}
			deskewLogf("[%s] line-angle rejected by projection consistency", inPath)
		}
	}

	bestA := 0.0
	bestS := -1.0
	searchMin := -p.AngleLimit
	searchMax := p.AngleLimit
	if lineOK {
		window := math.Max(1.5, 2*lineDispersion)
		searchMin = math.Max(-p.AngleLimit, lineAngle-window)
		searchMax = math.Min(p.AngleLimit, lineAngle+window)
	}

	for a := searchMin; a <= searchMax+1e-9; a += p.AngleStep {
		rot, err := rotateSameSize(inv, a)
		if err != nil {
			return 0, fmt.Errorf("rotate for angle %.2f: %w", a, err)
		}
		s := projectionVariance(rot)
		rot.Close()

		deskewLogf("[%s] angle=%.2f score=%.6f", inPath, a, s)

		if s > bestS {
			bestS = s
			bestA = a
		}
	}

	deskewLogf("[%s] best angle=%.3f score=%.6f", inPath, bestA, bestS)

	return bestA, nil
}

func refineProjectionAroundAngle(inv gocv.Mat, center float64, p projParams, inPath string) (float64, float64, float64, float64, error) {
	window := 0.75
	start := math.Max(-p.AngleLimit, center-window)
	end := math.Min(p.AngleLimit, center+window)

	bestA := 0.0
	bestS := -1.0
	minS := math.MaxFloat64
	for a := start; a <= end+1e-9; a += p.AngleStep {
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

func estimateSkewTextLines(invBinary gocv.Mat, p projParams, inPath string) (float64, float64, bool, error) {
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	if rows == 0 || cols == 0 {
		return 0, 0, false, nil
	}

	kernelW := cols / 25
	if kernelW < 15 {
		kernelW = 15
	}
	if kernelW > 63 {
		kernelW = 63
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelW, 3))
	defer kernel.Close()

	merged := gocv.NewMat()
	defer merged.Close()
	if err := gocv.MorphologyEx(invBinary, &merged, gocv.MorphClose, kernel); err != nil {
		return 0, 0, false, fmt.Errorf("merge text lines: %w", err)
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
		if math.Abs(observed) > p.AngleLimit {
			continue
		}

		weight := math.Sqrt(area) * float64(r.Dx())
		contourSamples = append(contourSamples, weightedAngleSample{
			angle:  observed,
			weight: weight,
		})
	}

	houghSamples, houghWeight, err := estimateSkewHoughLines(merged, p)
	if err != nil {
		return 0, 0, false, fmt.Errorf("estimate hough lines: %w", err)
	}

	totalWeight := sampleWeight(contourSamples) + houghWeight
	totalSamples := len(contourSamples) + len(houghSamples)
	if totalSamples < 3 || totalWeight < float64(cols)*10 {
		deskewLogf("[%s] line-angle unavailable samples=%d totalWeight=%.1f", inPath, totalSamples, totalWeight)
		return 0, 0, false, nil
	}

	contourCluster := summarizeAngleCluster("contours", contourSamples, p.AngleStep, 0.75)
	houghCluster := summarizeAngleCluster("hough", houghSamples, p.AngleStep, 0.75)
	combinedSamples := append(append([]weightedAngleSample(nil), contourSamples...), houghSamples...)
	combinedCluster := summarizeAngleCluster("combined", combinedSamples, p.AngleStep, 0.75)

	logAngleCluster(inPath, contourCluster)
	logAngleCluster(inPath, houghCluster)
	logAngleCluster(inPath, combinedCluster)

	best := chooseAngleCluster(contourCluster, houghCluster, combinedCluster)
	if !best.ok || best.inliers < 3 || best.inlierWeight < float64(cols)*6 {
		deskewLogf("[%s] line-angle rejected weak cluster source=%s peak=%.3f inliers=%d inlierWeight=%.1f", inPath, best.source, best.peak, best.inliers, best.inlierWeight)
		return 0, 0, false, nil
	}
	deskewLogf("[%s] line-angle chose source=%s median=%.3f dispersion=%.3f", inPath, best.source, best.median, best.dispersion)

	if best.dispersion > 1.25 {
		deskewLogf("[%s] line-angle rejected due to inlier dispersion %.3f", inPath, best.dispersion)
		return 0, best.dispersion, false, nil
	}

	return best.median, best.dispersion, true, nil
}

func estimateSkewHoughLines(merged gocv.Mat, p projParams) ([]weightedAngleSample, float64, error) {
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

	minLineLength := float32(cols) * 0.20
	maxLineGap := float32(cols) * 0.04
	if err := gocv.HoughLinesPWithParams(merged, &lines, 1, math.Pi/180, threshold, minLineLength, maxLineGap); err != nil {
		return nil, 0, err
	}
	if lines.Empty() {
		return nil, 0, nil
	}

	samples := make([]weightedAngleSample, 0, lines.Rows())
	totalWeight := 0.0
	maxAngle := p.AngleLimit

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
		if tiltWeight < 0.35 {
			tiltWeight = 0.35
		}

		samples = append(samples, weightedAngleSample{
			angle:  angle,
			weight: length * length * tiltWeight,
		})
		totalWeight += length * length * tiltWeight
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
	for _, cluster := range clusters {
		if !cluster.ok {
			continue
		}

		score := cluster.inlierWeight / (1 + cluster.dispersion)
		if math.Abs(cluster.median) < 0.2 {
			score *= 0.75
		}
		if math.Abs(cluster.median) >= 0.4 {
			score *= 1.15
		}

		if score > bestScore {
			best = cluster
			bestScore = score
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
	// invBinary: ink pixels are >0, background 0.
	// Score = variance of row sums (peaky rows = better alignment).
	rows := invBinary.Rows()
	cols := invBinary.Cols()
	if rows == 0 || cols == 0 {
		return 0
	}

	sums := make([]float64, rows)

	for y := 0; y < rows; y++ {
		row := invBinary.RowRange(y, y+1)
		// Count non-zero pixels
		nz := gocv.CountNonZero(row)
		row.Close()
		sums[y] = float64(nz) / float64(cols)
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
