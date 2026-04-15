//go:build !nogocv

package formatcov

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"gocv.io/x/gocv"
)

const deskewAngleTolerance = 0.3

var expectedDeskewAngles = map[string]float64{
	"Antwerp_1654_025.snap.png":    0,
	"Antwerp_1654_026.snap.png":    0.3,
	"Antwerp_1654_198.snap.png":    0,
	"Antwerp_1654_222.snap.png":    0,
	"Basel_1533_007.snap.png":      0,
	"Basel_1533_015.snap.png":      -1.0,
	"Basel_1533_036.snap.png":      0,
	"Leiden_1606_045.snap.png":     0.6,
	"Livorno_1709_019.snap.png":    0.85,
	"Livorno_1709_064.snap.png":    -1.5,
	"London_1570_033.snap.png":     -0.78,
	"London_1570_035.snap.png":     -0.4,
	"London_1570_395.snap.png":     -0.8,
	"London_1570_493.snap.png":     -1.66,
	"Lyon_1603_18.snap.png":        0,
	"Lyon_1603_24.snap.png":        0.6,
	"Lyon_1603_25.snap.png":        0.5,
	"Lyon_1603_90.snap.png":        -4,
	"Oxford_1703_111.snap.png":     -0.6,
	"Oxford_1703_263.snap.png":     0.18,
	"Oxford_1703_626.snap.png":     0,
	"Oxford_1705_276.snap.png":     0,
	"Oxford_1705_289.snap.png":     0.9,
	"Paris_1566_094.snap.png":      -0.37,
	"Paris_1615_008.snap.png":      0.5,
	"Paris_1615_009.snap.png":      -0.7,
	"Paris_1615_010.snap.png":      -0.5,
	"Paris_1615_047.snap.png":      0,
	"Paris_1615_149.snap.png":      -0.84,
	"Paris_1634_vol4_193.snap.png": 1.1,
	"Paris_1634_vol5_004.snap.png": 1.6,
	"Paris_1634_vol5_642.snap.png": 1.6,
	"Paris_1634_vol5_791.snap.png": -0.75,
	"Paris_1667_023.snap.png":      0,
	"Paris_1667_069.snap.png":      0,
	"Paris_1667_307.snap.png":      0,
	"Paris_1682a_348.snap.png":     0.4,
	"Paris_1682a_463.snap.png":     0,
	"Paris_1682a_481.snap.png":     -1,
	"Paris_1682a_668.snap.png":     0,
	"Paris_1685_273.snap.png":      0.3,
	"Paris_1814_vol3_135.snap.png": -0.55,
	"Pesaro_1572_054.snap.png":     0,
	"Pesaro_1572_181.snap.png":     0,
	"Pesaro_1572_229.snap.png":     -0.7,
	"Pesaro_1572_321.snap.png":     0,
	"Rome_1574_171.snap.png":       0,
	"Rome_1680_015.snap.png":       0,
	"Rome_1680_017.snap.png":       0,
	"Seville_1576_024.snap.png":    0,
	"Seville_1576_114.snap.png":    0,
	"Seville_1576_115.snap.png":    -0.62,
	"Strasbourg_1566_026.snap.png": -0.6,
	"Strasbourg_1566_053.snap.png": 0.32,
	"Strasbourg_1566_117.snap.png": 0,
	"The_Hague_1758_075.snap.png":  0,
	"Venice_1482_018.snap.png":     1.6,
	"Vienna_1694_092.snap.png":     -1.8,
	"Vienna_1694_188.snap.png":     -2.35,
}

func TestDeskewSnapAngles(t *testing.T) {
	paths, err := deskewFixtureInputs(filepath.Join("testdata", "denoise"))
	if err != nil {
		t.Fatalf("list deskew fixtures: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no deskew fixture inputs found")
	}

	sem := make(chan struct{}, runtime.NumCPU())

	for _, inputPath := range paths {
		inputPath := inputPath
		t.Run(strings.TrimSuffix(filepath.Base(inputPath), ".snap.png"), func(t *testing.T) {
			t.Parallel()

			sem <- struct{}{}
			defer func() { <-sem }()

			angle, err := estimateDeskewAngleForPath(inputPath)
			if err != nil {
				t.Fatalf("estimate deskew angle for %q: %v", inputPath, err)
			}

			name := filepath.Base(inputPath)
			expected, ok := expectedDeskewAngles[name]
			if !ok {
				t.Fatalf("missing expected angle for %q: actual=%.3f", name, angle)
			}

			if math.Abs(angle-expected) > deskewAngleTolerance {
				t.Fatalf("deskew angle mismatch for %q: actual=%.3f expected=%.3f tolerance=%.3f", name, angle, expected, deskewAngleTolerance)
			}
		})
	}
}

func deskewFixtureInputs(dir string) ([]string, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.snap.png"))
	if err != nil {
		return nil, err
	}
	if paths == nil {
		return []string{}, nil
	}
	slices.Sort(paths)
	return paths, nil
}

func estimateDeskewAngleForPath(path string) (float64, error) {
	img := gocv.IMRead(path, gocv.IMReadColor)
	if img.Empty() {
		return 0, fmt.Errorf("read image %q: empty or unsupported format", path)
	}
	defer img.Close()

	small := img.Clone()
	if deskewDownscaleMax > 0 {
		var err error
		small, err = resizeMaxSide(small, deskewDownscaleMax)
		if err != nil {
			return 0, fmt.Errorf("downscale for angle estimation: %w", err)
		}
	}
	defer small.Close()

	angle, err := estimateSkewProjection(small, path)
	if err != nil {
		return 0, err
	}
	return angle, nil
}
