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
	denoiseMinWorkers = 1
	denoiseMaxWorkers = 2

	denoiseOneSCurveSteepness      = 12.0
	denoiseOneAlphaScale           = 1
	denoiseOneAlphaMinThreshold    = 0
	denoiseOneSmallBlobFraction    = 0.00005
	denoiseOneRemoteRadiusFraction = 0.08
	denoiseOneBlobThreshGray       = 200.0
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
	applySCurve(&gray, &contrastA)

	alphaGrayThresh := float32(255.0 * (1.0 - denoiseOneAlphaMinThreshold/denoiseOneAlphaScale))

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

	if ok := gocv.IMWrite(outPath, filtered); !ok {
		return fmt.Errorf("write image %q", outPath)
	}
	return nil
}

func applySCurve(src, dst *gocv.Mat) {
	lut := gocv.NewMatWithSize(1, 256, gocv.MatTypeCV8U)
	defer lut.Close()
	lutData, _ := lut.DataPtrUint8()
	yMin := 1.0 / (1.0 + math.Exp(denoiseOneSCurveSteepness*0.5))
	yMax := 1.0 / (1.0 + math.Exp(-denoiseOneSCurveSteepness*0.5))
	for i := 0; i < 256; i++ {
		x := float64(i)/255.0 - 0.5
		y := 1.0 / (1.0 + math.Exp(-denoiseOneSCurveSteepness*x))
		norm := (y - yMin) / (yMax - yMin)
		lutData[i] = uint8(math.Round(norm * 255))
	}
	gocv.LUT(*src, lut, dst)
}

func removeSmallIsolatedBlobs(img *gocv.Mat, minDim float64) {
	h, w := img.Rows(), img.Cols()
	totalArea := float64(h * w)
	smallBlobMaxArea := denoiseOneSmallBlobFraction * totalArea
	remoteRadiusSq := math.Pow(denoiseOneRemoteRadiusFraction*minDim, 2)

	binary := gocv.NewMat()
	defer binary.Close()
	gocv.Threshold(*img, &binary, float32(denoiseOneBlobThreshGray), 255, gocv.ThresholdBinaryInv)

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
