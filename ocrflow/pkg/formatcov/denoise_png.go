//go:build !nogocv

package formatcov

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gocv.io/x/gocv"
	"golang.org/x/sync/errgroup"
)

const (
	// Background normalization kernel — must span inter-character gaps.
	denoiseBackgroundKernelSize = 51

	// Adaptive threshold parameters.
	denoiseAdaptiveBlockSize = 61 // must be odd
	denoiseAdaptiveC         = 20 // higher = reject more near-background pixels

	// Morphological opening after binarization — removes thin noise strands.
	denoiseMorphOpenSize = 2

	// Second, larger closing+opening pass to kill remaining diffuse speckles.
	denoiseSpeckleKillOpenSize  = 4 // increased: erodes away more speckle clusters
	denoiseSpeckleKillCloseSize = 2

	// Contour-based speckle filter.
	denoiseSpeckleMinArea   = 1.0
	denoiseSpeckleMaxArea   = 55.0 // wider: catch larger bleed-through clusters
	denoiseSpeckleMaxWidth  = 12
	denoiseSpeckleMaxHeight = 12
	denoiseSpeckleMinFill   = 0.20
	denoiseSpeckleMinAspect = 0.30
	denoiseSpeckleMaxAspect = 3.0

	// Connected-component minimum pixel area — blobs smaller than this are
	// unconditionally removed regardless of shape.
	// Raised from 5: real text strokes are always larger than this.
	denoiseMinBlobArea = 12

	denoiseMaskColor  = uint8(255)
	denoiseMinWorkers = 2
	denoiseMaxWorkers = 24
)

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
//  8. Invert to black-on-white output
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

	// --- 2. Background normalization ---
	bgKernel := gocv.GetStructuringElement(
		gocv.MorphRect,
		image.Point{X: denoiseBackgroundKernelSize, Y: denoiseBackgroundKernelSize},
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
		image.Point{X: denoiseMorphOpenSize, Y: denoiseMorphOpenSize},
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
		image.Point{X: denoiseSpeckleKillOpenSize, Y: denoiseSpeckleKillOpenSize},
	)
	defer k2open.Close()

	opened2 := gocv.NewMat()
	defer opened2.Close()
	gocv.MorphologyEx(opened1, &opened2, gocv.MorphOpen, k2open)

	k2close := gocv.GetStructuringElement(
		gocv.MorphEllipse,
		image.Point{X: denoiseSpeckleKillCloseSize, Y: denoiseSpeckleKillCloseSize},
	)
	defer k2close.Close()

	closed2 := gocv.NewMat()
	defer closed2.Close()
	gocv.MorphologyEx(opened2, &closed2, gocv.MorphClose, k2close)

	// --- 6. Connected-component minimum-size filter ---
	// Unconditionally removes every blob smaller than denoiseMinBlobArea pixels,
	// regardless of shape. This catches the finest residual speckle dust.
	afterCC, err := removeSmallBlobs(closed2, denoiseMinBlobArea)
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

		if w > denoiseSpeckleMaxWidth ||
			h > denoiseSpeckleMaxHeight ||
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

	// --- 8. Invert to black-on-white ---
	out := gocv.NewMat()
	defer out.Close()
	gocv.BitwiseNot(cleaned, &out)

	if ok := gocv.IMWrite(outPath, out); !ok {
		return fmt.Errorf("write image %q: %w", outPath, errWriteFailed)
	}
	return nil
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
