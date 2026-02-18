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
	"strings"

	"github.com/samber/lo"
	"gocv.io/x/gocv"
	"golang.org/x/sync/errgroup"
)

var errWriteFailed = errors.New("gocv IMWrite failed")

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
		AngleLimit:   6.0,
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

	angle, err := estimateSkewProjection(small, p)
	if err != nil {
		return err
	}

	if math.Abs(angle) < p.MinRotate {
		// Write original unchanged (still useful for comparison)
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

func estimateSkewProjection(img gocv.Mat, p projParams) (float64, error) {
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

	bestA := 0.0
	bestS := -1.0

	for a := -p.AngleLimit; a <= p.AngleLimit+1e-9; a += p.AngleStep {
		rot, err := rotateKeepAll(inv, a, 0)
		if err != nil {
			return 0, fmt.Errorf("rotate for angle %.2f: %w", a, err)
		}
		s := projectionVariance(rot)
		rot.Close()

		if s > bestS {
			bestS = s
			bestA = a
		}
	}

	return bestA, nil
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
			c.Close()
			continue
		}
		r := gocv.BoundingRect(c)
		c.Close()
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
