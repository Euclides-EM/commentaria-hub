package formatcov

import (
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/testutils"
	"os"
	"path/filepath"
	"testing"
)

func TestAlto2Yolo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yolo2alto_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origAltoDir := filepath.Join(tmpDir, "orig_alto")
	yoloDir1 := filepath.Join(tmpDir, "yolo_converted")
	altoDir1 := filepath.Join(tmpDir, "alto_converted")
	yoloDir2 := filepath.Join(tmpDir, "yolo_converted2")
	altoDir2 := filepath.Join(tmpDir, "alto_converted2")

	// start with ALTO files
	if err := os.MkdirAll(origAltoDir, 0o755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	testFilesToCopy := []string{
		filepath.Join("testdata", "alto", "page-0315.xml"),
		filepath.Join("testdata", "alto", "page-0399.xml"),
	}
	for _, file := range testFilesToCopy {
		destPath := filepath.Join(origAltoDir, filepath.Base(file))
		if err := futils.CopyFile(file, destPath); err != nil {
			t.Fatalf("failed to copy test file %s to %s: %v", file, destPath, err)
		}
	}

	// convert the files to YOLO
	if err := Alto2Yolo(filepath.Join("testdata", "img"), origAltoDir, yoloDir1, 0.0, ""); err != nil {
		t.Fatalf("Alto2Yolo failed: %v", err)
	}

	// convert back to ALTO
	// this ALTO is different from the original one, since some info is lost in YOLO format
	if err := Yolo2Alto(yoloDir1, altoDir1); err != nil {
		t.Fatalf("Yolo2Alto failed: %v", err)
	}

	// convert back to YOLO
	// this YOLO should be the same as the first converted YOLO
	if err := Alto2Yolo(filepath.Join("testdata", "img"), altoDir1, yoloDir2, 0.0, ""); err != nil {
		t.Fatalf("Alto2Yolo failed: %v", err)
	}
	if err := testutils.CompareDirs(yoloDir1, yoloDir2, map[string]string{yoloDir1: yoloDir2}); err != nil {
		t.Fatalf("Converted YOLO dirs do not match: %v", err)
	}

	// convert back to ALTO to check consistency
	// this ALTO should be the same as the first converted ALTO
	if err := Yolo2Alto(yoloDir2, altoDir2); err != nil {
		t.Fatalf("Yolo2Alto failed: %v", err)
	}
	if err := testutils.CompareDirs(altoDir1, altoDir2, map[string]string{}); err != nil {
		t.Fatalf("Converted ALTO dirs do not match: %v", err)
	}

}
