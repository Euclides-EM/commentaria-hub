//go:build !nogocv

package formatcov

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestDenoisePNGFixtures(t *testing.T) {
	paths, err := denoiseFixtureInputs("testdata/denoise")
	if err != nil {
		t.Fatalf("list denoise fixtures: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no denoise fixture inputs found")
	}

	sem := make(chan struct{}, runtime.NumCPU())

	for _, inputPath := range paths {
		inputPath := inputPath
		t.Run(strings.TrimSuffix(filepath.Base(inputPath), ".png"), func(t *testing.T) {
			t.Parallel()

			sem <- struct{}{}
			defer func() { <-sem }()

			stagePath := strings.TrimSuffix(inputPath, ".png") + ".stage.png"
			snapPath := strings.TrimSuffix(inputPath, ".png") + ".snap.png"

			if err := denoiseOne(inputPath, stagePath); err != nil {
				t.Fatalf("denoise %q: %v", inputPath, err)
			}

			if _, err := os.Stat(snapPath); err != nil {
				if os.IsNotExist(err) {
					t.Fatalf("missing snapshot for %q: %s", inputPath, snapPath)
				}
				t.Fatalf("stat snapshot %q: %v", snapPath, err)
			}

			matchRatio, err := comparePNGMatchRatio(snapPath, stagePath)
			if err != nil {
				t.Fatalf("compare snapshot %q to stage %q: %v", snapPath, stagePath, err)
			}
			if matchRatio < 0.98 {
				t.Fatalf("snapshot mismatch for %q: %.2f%% pixels matched, need at least 98.00%%", inputPath, matchRatio*100)
			}
		})
	}
}

func denoiseFixtureInputs(dir string) ([]string, error) {
	paths := []string{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if filepath.Ext(name) != ".png" {
			return nil
		}
		if strings.HasSuffix(name, ".snap.png") || strings.HasSuffix(name, ".stage.png") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.Sort(paths)
	return paths, nil
}

func comparePNGMatchRatio(pathA, pathB string) (float64, error) {
	imgA, err := decodePNG(pathA)
	if err != nil {
		return 0, err
	}
	imgB, err := decodePNG(pathB)
	if err != nil {
		return 0, err
	}

	boundsA := imgA.Bounds()
	boundsB := imgB.Bounds()
	if boundsA.Dx() != boundsB.Dx() || boundsA.Dy() != boundsB.Dy() {
		return 0, fmt.Errorf("image size mismatch: %s is %dx%d, %s is %dx%d", pathA, boundsA.Dx(), boundsA.Dy(), pathB, boundsB.Dx(), boundsB.Dy())
	}

	total := boundsA.Dx() * boundsA.Dy()
	if total == 0 {
		return 1, nil
	}

	matched := 0
	for y := 0; y < boundsA.Dy(); y++ {
		for x := 0; x < boundsA.Dx(); x++ {
			if samePixel(imgA, boundsA.Min.X+x, boundsA.Min.Y+y, imgB, boundsB.Min.X+x, boundsB.Min.Y+y) {
				matched++
			}
		}
	}

	return float64(matched) / float64(total), nil
}

func decodePNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func samePixel(imgA image.Image, ax, ay int, imgB image.Image, bx, by int) bool {
	r1, g1, b1, a1 := imgA.At(ax, ay).RGBA()
	r2, g2, b2, a2 := imgB.At(bx, by).RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
