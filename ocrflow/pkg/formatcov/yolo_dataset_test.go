package formatcov

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateYoloDatasetLayouts(t *testing.T) {
	t.Parallel()

	for _, subDir := range []string{"", "train", "val", "valid", "test"} {
		subDir := subDir
		t.Run(subDir, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "labelmap.txt"), []byte("MainZone\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(dir, subDir, "labels"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(dir, subDir, "images"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, subDir, "labels", "page-0001.txt"), nil, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, subDir, "images", "page-0001.PNG"), nil, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := ValidateYoloDataset(dir); err != nil {
				t.Fatalf("valid %q layout rejected: %v", subDir, err)
			}
		})
	}
}

func TestValidateYoloDatasetRejectsMissingImage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "labelmap.txt"), []byte("MainZone\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "labels"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "labels", "page-0001.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateYoloDataset(dir); err == nil {
		t.Fatal("expected missing image to fail validation")
	}
}

func TestValidateYoloDatasetRejectsImageOutsideBundle(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	dir := filepath.Join(parent, "yolo")
	if err := os.MkdirAll(filepath.Join(dir, "labels"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(parent, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "labelmap.txt"), []byte("MainZone\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "labels", "page-0001.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "images", "page-0001.jpg"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateYoloDataset(dir); err == nil {
		t.Fatal("expected an image outside the YOLO bundle to fail validation")
	}
}

func TestYolo2AltoUsesImageFromValSplit(t *testing.T) {
	t.Parallel()

	yoloDir := t.TempDir()
	altoDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(yoloDir, "labelmap.txt"), []byte("MainZone\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(yoloDir, "val", "labels"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(yoloDir, "val", "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(yoloDir, "val", "labels", "page-0001.txt"), []byte("0 0.5 0.5 0.5 0.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	imagePath := filepath.Join(yoloDir, "val", "images", "page-0001.png")
	f, err := os.Create(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 20, 10))
	img.Set(0, 0, color.White)
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := Yolo2Alto(yoloDir, altoDir); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(altoDir, "page-0001.xml"))
	if err != nil {
		t.Fatal(err)
	}
	xml := string(raw)
	for _, want := range []string{"<fileName>page-0001.png</fileName>", `WIDTH="20"`, `HEIGHT="10"`} {
		if !strings.Contains(xml, want) {
			t.Fatalf("converted ALTO does not contain %q", want)
		}
	}
}

func TestValidateYoloImagesAgainstDataset(t *testing.T) {
	t.Parallel()

	t.Run("matching dimensions", func(t *testing.T) {
		yoloDir, datasetImageDir := createGeometryValidationFixture(t, 20, 10, 20, 10)
		if err := ValidateYoloImagesAgainstDataset(yoloDir, datasetImageDir); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("resized export", func(t *testing.T) {
		yoloDir, datasetImageDir := createGeometryValidationFixture(t, 10, 5, 20, 10)
		err := ValidateYoloImagesAgainstDataset(yoloDir, datasetImageDir)
		if err == nil {
			t.Fatal("expected resized YOLO image to be rejected")
		}
		if !strings.Contains(err.Error(), "geometrically transformed YOLO imports are not supported") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func createGeometryValidationFixture(t *testing.T, yoloWidth, yoloHeight, datasetWidth, datasetHeight int) (string, string) {
	t.Helper()
	yoloDir := t.TempDir()
	datasetImageDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(yoloDir, "labelmap.txt"), []byte("MainZone\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(yoloDir, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(yoloDir, "labels"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(yoloDir, "labels", "page-0001_png.rf.abc123.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	writeTestPNG(t, filepath.Join(yoloDir, "images", "page-0001_png.rf.abc123.png"), yoloWidth, yoloHeight)
	writeTestPNG(t, filepath.Join(datasetImageDir, "page-0001.png"), datasetWidth, datasetHeight)
	return yoloDir, datasetImageDir
}

func writeTestPNG(t *testing.T, path string, width int, height int) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
