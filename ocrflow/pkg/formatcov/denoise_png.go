//go:build !nogocv

package formatcov

import (
	"context"
	"fmt"
	"image"
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
	denoiseMinWorkers = 1
	denoiseMaxWorkers = 2
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

func DenoisePNGFile(inPath, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create output dir %q: %w", filepath.Dir(outPath), err)
	}
	if err := denoiseOne(inPath, outPath); err != nil {
		return fmt.Errorf("denoise %q: %w", inPath, err)
	}
	return nil
}

func denoiseOne(inPath, outPath string) error {
	const initialSCurveSteepness = 6.0

	img := gocv.IMRead(inPath, gocv.IMReadColor)
	if img.Empty() {
		return fmt.Errorf("read image %q: empty or unsupported format", inPath)
	}
	defer img.Close()

	h, w := img.Rows(), img.Cols()
	minDim := float64(w)
	if h < w {
		minDim = float64(h)
	}

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

	contrastA := gocv.NewMat()
	defer contrastA.Close()
	applySCurve(&gray, &contrastA, initialSCurveSteepness)

	alphaGrayThresh := float32(255.0 * (1.0 - 0.15/1.0))

	maskB := gocv.NewMat()
	defer maskB.Close()
	gocv.Threshold(gray, &maskB, alphaGrayThresh, 255, gocv.ThresholdBinaryInv)

	combined := gocv.NewMatWithSize(h, w, gocv.MatTypeCV8U)
	defer combined.Close()
	combined.SetTo(gocv.NewScalar(255, 0, 0, 0))
	if err := contrastA.CopyToWithMask(&combined, maskB); err != nil {
		return fmt.Errorf("combine tracks: %w", err)
	}

	filtered := combined.Clone()
	defer filtered.Close()
	removeSmallIsolatedBlobs(&filtered, minDim)
	removeSpecks(&filtered, minDim)
	applyFinalToneMapping(&filtered)

	if ok := gocv.IMWrite(outPath, filtered); !ok {
		return fmt.Errorf("write image %q", outPath)
	}
	return nil
}

func applyFinalToneMapping(img *gocv.Mat) {
	const (
		darkPixelThresh      = 230
		brightnessBoostValue = 80
		finalDarkSCurve      = 12.0
		focusFraction        = 0.7
	)

	imgData, err := img.DataPtrUint8()
	if err != nil || len(imgData) == 0 {
		return
	}

	focusData := imgData
	rows, cols := img.Rows(), img.Cols()
	if rows > 0 && cols > 0 {
		focusWidth := int(math.Round(float64(cols) * focusFraction))
		focusHeight := int(math.Round(float64(rows) * focusFraction))
		if focusWidth < 1 {
			focusWidth = 1
		}
		if focusHeight < 1 {
			focusHeight = 1
		}

		startX := (cols - focusWidth) / 2
		startY := (rows - focusHeight) / 2
		focusRegion := img.Region(image.Rect(startX, startY, startX+focusWidth, startY+focusHeight))
		defer focusRegion.Close()

		focusClone := focusRegion.Clone()
		defer focusClone.Close()

		if regionData, regionErr := focusClone.DataPtrUint8(); regionErr == nil && len(regionData) > 0 {
			focusData = regionData
		}
	}

	darkCount := 0
	for _, px := range focusData {
		if px < darkPixelThresh {
			darkCount++
		}
	}

	ratio := float64(darkCount) / float64(len(focusData))
	if ratio > 0.9 {
		applyDarkenNonBrightPixels(imgData, 0.1, 30)
	}
	if ratio < 0.25 {
		applyDarkenNonBrightPixels(imgData, 0.25, 100)
		return
	} else if ratio < 0.8 {
		return
	}

	for i, px := range imgData {
		boosted := int(px) + brightnessBoostValue
		if boosted > 255 {
			boosted = 255
		}
		imgData[i] = uint8(boosted)
	}

	recurved := gocv.NewMat()
	defer recurved.Close()
	applySCurve(img, &recurved, finalDarkSCurve)
	recurved.CopyTo(img)
}

func applyDarkenNonBrightPixels(imgData []uint8, brightKeepPercentile float64, brightDarkenValue int) {
	brightKeepThreshold := percentileNonBrightPixelValue(imgData, brightKeepPercentile)

	for i, px := range imgData {
		if px >= brightKeepThreshold {
			continue
		}
		darkened := int(px) - brightDarkenValue
		if darkened < 0 {
			darkened = 0
		}
		imgData[i] = uint8(darkened)
	}
}

func percentileNonBrightPixelValue(imgData []uint8, percentile float64) uint8 {
	var counts [256]int
	nonBrightCount := 0
	for _, px := range imgData {
		if px == 255 {
			continue
		}
		counts[px]++
		nonBrightCount++
	}
	if nonBrightCount == 0 {
		return 255
	}

	if percentile < 0 {
		percentile = 0
	} else if percentile > 1 {
		percentile = 1
	}

	target := int(math.Ceil(percentile*float64(nonBrightCount))) - 1
	if target < 0 {
		target = 0
	}

	seen := 0
	for value, count := range counts {
		seen += count
		if seen > target {
			return uint8(value)
		}
	}

	return 255
}

func applySCurve(src, dst *gocv.Mat, steepness float64) {
	lut := gocv.NewMatWithSize(1, 256, gocv.MatTypeCV8U)
	defer lut.Close()
	lutData, _ := lut.DataPtrUint8()
	yMin := 1.0 / (1.0 + math.Exp(steepness*0.5))
	yMax := 1.0 / (1.0 + math.Exp(-steepness*0.5))
	for i := 0; i < 256; i++ {
		x := float64(i)/255.0 - 0.5
		y := 1.0 / (1.0 + math.Exp(-steepness*x))
		norm := (y - yMin) / (yMax - yMin)
		lutData[i] = uint8(math.Round(norm * 255))
	}
	gocv.LUT(*src, lut, dst)
}

func removeSpecks(img *gocv.Mat, minDim float64) {
	const (
		speckMaxFraction           = 0.00003
		speckProximityFraction     = 0.0005
		speckAspectRatioMax        = 2.5
		speckFillRatioMin          = 0.3
		speckLightFraction         = 0.90
		speckClusterRadiusFraction = 0.01
		speckTextProximityFraction = 0.02
		speckHaloFraction          = 0.003
		oneSpeckLightThresh        = 200.0
	)

	h, w := img.Rows(), img.Cols()
	totalArea := float64(h * w)
	speckMaxArea := speckMaxFraction * totalArea
	proximityRadius := int(math.Round(speckProximityFraction * minDim))
	if proximityRadius < 1 {
		proximityRadius = 1
	}

	binary := gocv.NewMat()
	defer binary.Close()
	gocv.Threshold(*img, &binary, float32(oneSpeckLightThresh), 255, gocv.ThresholdBinaryInv)

	labels := gocv.NewMat()
	defer labels.Close()
	stats := gocv.NewMat()
	defer stats.Close()
	centroids := gocv.NewMat()
	defer centroids.Close()

	n := gocv.ConnectedComponentsWithStats(binary, &labels, &stats, &centroids)
	if n <= 1 {
		return
	}

	imgData, err := img.DataPtrUint8()
	if err != nil {
		return
	}

	type centroid struct{ cx, cy float64 }
	centroids2 := make([]centroid, n)
	for i := 1; i < n; i++ {
		centroids2[i] = centroid{
			cx: centroids.GetDoubleAt(i, 0),
			cy: centroids.GetDoubleAt(i, 1),
		}
	}

	clusterRadiusSq := math.Pow(speckClusterRadiusFraction*minDim, 2)

	isCandidate := make([]bool, n)
	for i := 1; i < n; i++ {
		area := int(stats.GetIntAt(i, 4))
		if float64(area) > speckMaxArea {
			continue
		}
		bw := int(stats.GetIntAt(i, 2))
		bh := int(stats.GetIntAt(i, 3))

		aspectRatio := float64(bw) / float64(bh)
		if aspectRatio > speckAspectRatioMax || aspectRatio < 1.0/speckAspectRatioMax {
			continue
		}
		fillRatio := float64(area) / float64(bw*bh)
		if fillRatio < speckFillRatioMin {
			continue
		}

		x0 := int(stats.GetIntAt(i, 0))
		y0 := int(stats.GetIntAt(i, 1))
		x1 := x0 + bw - 1
		y1 := y0 + bh - 1

		px0 := x0 - proximityRadius
		if px0 < 0 {
			px0 = 0
		}
		py0 := y0 - proximityRadius
		if py0 < 0 {
			py0 = 0
		}
		px1 := x1 + proximityRadius
		if px1 >= w {
			px1 = w - 1
		}
		py1 := y1 + proximityRadius
		if py1 >= h {
			py1 = h - 1
		}

		lightCount := 0
		totalCount := 0
		for ry := py0; ry <= py1; ry++ {
			for rx := px0; rx <= px1; rx++ {
				if int(labels.GetIntAt(ry, rx)) == i {
					continue
				}
				totalCount++
				if imgData[ry*w+rx] >= uint8(oneSpeckLightThresh) {
					lightCount++
				}
			}
		}

		if totalCount == 0 {
			continue
		}
		if float64(lightCount)/float64(totalCount) >= speckLightFraction {
			isCandidate[i] = true
		}
	}

	textProximityRadius := int(math.Round(speckTextProximityFraction * minDim))

	toRemove := make([]bool, n)
	for i := 1; i < n; i++ {
		if !isCandidate[i] {
			continue
		}
		ix0 := int(stats.GetIntAt(i, 0)) - textProximityRadius
		iy0 := int(stats.GetIntAt(i, 1)) - textProximityRadius
		ix1 := int(stats.GetIntAt(i, 0)) + int(stats.GetIntAt(i, 2)) - 1 + textProximityRadius
		iy1 := int(stats.GetIntAt(i, 1)) + int(stats.GetIntAt(i, 3)) - 1 + textProximityRadius

		hasNeighbor := false
		nearText := false
		for j := 1; j < n; j++ {
			if j == i {
				continue
			}
			if isCandidate[j] {
				dx := centroids2[i].cx - centroids2[j].cx
				dy := centroids2[i].cy - centroids2[j].cy
				if dx*dx+dy*dy <= clusterRadiusSq {
					hasNeighbor = true
				}
			} else {
				jx0 := int(stats.GetIntAt(j, 0))
				jy0 := int(stats.GetIntAt(j, 1))
				jx1 := jx0 + int(stats.GetIntAt(j, 2)) - 1
				jy1 := jy0 + int(stats.GetIntAt(j, 3)) - 1
				if ix0 <= jx1 && ix1 >= jx0 && iy0 <= jy1 && iy1 >= jy0 {
					nearText = true
				}
			}
			if hasNeighbor && nearText {
				break
			}
		}
		if !hasNeighbor && !nearText {
			toRemove[i] = true
		}
	}

	haloRadius := int(math.Round(speckHaloFraction * minDim))
	if haloRadius < 1 {
		haloRadius = 1
	}

	for i := 1; i < n; i++ {
		if !toRemove[i] {
			continue
		}
		x0 := int(stats.GetIntAt(i, 0))
		y0 := int(stats.GetIntAt(i, 1))
		bw := int(stats.GetIntAt(i, 2))
		bh := int(stats.GetIntAt(i, 3))

		ex0 := x0 - haloRadius
		if ex0 < 0 {
			ex0 = 0
		}
		ey0 := y0 - haloRadius
		if ey0 < 0 {
			ey0 = 0
		}
		ex1 := x0 + bw - 1 + haloRadius
		if ex1 >= w {
			ex1 = w - 1
		}
		ey1 := y0 + bh - 1 + haloRadius
		if ey1 >= h {
			ey1 = h - 1
		}

		for ry := ey0; ry <= ey1; ry++ {
			for rx := ex0; rx <= ex1; rx++ {
				lbl := int(labels.GetIntAt(ry, rx))
				if lbl == i || lbl == 0 || toRemove[lbl] {
					imgData[ry*w+rx] = 255
				}
			}
		}
	}
}

func removeSmallIsolatedBlobs(img *gocv.Mat, minDim float64) {
	const (
		smallBlobFraction    = 0.00005
		remoteRadiusFraction = 0.08
		blobThreshGray       = 200.0
	)

	h, w := img.Rows(), img.Cols()
	totalArea := float64(h * w)
	smallBlobMaxArea := smallBlobFraction * totalArea
	remoteRadiusSq := math.Pow(remoteRadiusFraction*minDim, 2)

	binary := gocv.NewMat()
	defer binary.Close()
	gocv.Threshold(*img, &binary, float32(blobThreshGray), 255, gocv.ThresholdBinaryInv)

	labels := gocv.NewMat()
	defer labels.Close()
	stats := gocv.NewMat()
	defer stats.Close()
	centroids := gocv.NewMat()
	defer centroids.Close()

	n := gocv.ConnectedComponentsWithStats(binary, &labels, &stats, &centroids)
	if n <= 1 {
		return
	}

	type blobInfo struct {
		cx, cy float64
		small  bool
	}
	blobs := make([]blobInfo, n)
	for i := 1; i < n; i++ {
		area := stats.GetIntAt(i, 4)
		blobs[i] = blobInfo{
			cx:    centroids.GetDoubleAt(i, 0),
			cy:    centroids.GetDoubleAt(i, 1),
			small: float64(area) < smallBlobMaxArea,
		}
	}

	toRemove := make([]bool, n)
	for i := 1; i < n; i++ {
		if !blobs[i].small {
			continue
		}
		remote := true
		for j := 1; j < n; j++ {
			if j == i {
				continue
			}
			dx := blobs[i].cx - blobs[j].cx
			dy := blobs[i].cy - blobs[j].cy
			if dx*dx+dy*dy <= remoteRadiusSq {
				remote = false
				break
			}
		}
		if remote {
			toRemove[i] = true
		}
	}

	imgData, err := img.DataPtrUint8()
	if err != nil {
		return
	}

	for i := 1; i < n; i++ {
		if !toRemove[i] {
			continue
		}
		x0 := int(stats.GetIntAt(i, 0))
		y0 := int(stats.GetIntAt(i, 1))
		bw := int(stats.GetIntAt(i, 2))
		bh := int(stats.GetIntAt(i, 3))
		for ry := y0; ry < y0+bh; ry++ {
			for rx := x0; rx < x0+bw; rx++ {
				if int(labels.GetIntAt(ry, rx)) == i {
					imgData[ry*w+rx] = 255
				}
			}
		}
	}
}
