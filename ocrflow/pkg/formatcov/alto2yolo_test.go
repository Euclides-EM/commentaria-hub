package formatcov

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetAltoImageFileNameRebasesGPUFarmPath(t *testing.T) {
	t.Parallel()

	altoPath := filepath.Join(t.TempDir(), "page-0008.xml")
	input := `<alto><Description><sourceImageInformation><fileName>/pbs/home/user/jobs/detect_annotation/run-1/assets/images/page-0008.png</fileName></sourceImageInformation></Description></alto>`
	if err := os.WriteFile(altoPath, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := setAltoImageFileName(altoPath, "page-0008.png"); err != nil {
		t.Fatalf("setAltoImageFileName() error = %v", err)
	}

	got, err := os.ReadFile(altoPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "/pbs/home/") {
		t.Fatalf("stale GPU-farm path remains in ALTO: %s", got)
	}
	if !strings.Contains(string(got), "<fileName>page-0008.png</fileName>") {
		t.Fatalf("ALTO does not reference staged image: %s", got)
	}
}

func TestSetAltoImageFileNameRejectsMissingFileName(t *testing.T) {
	t.Parallel()

	altoPath := filepath.Join(t.TempDir(), "page-0008.xml")
	if err := os.WriteFile(altoPath, []byte(`<alto/>`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := setAltoImageFileName(altoPath, "page-0008.png")
	if err == nil || !strings.Contains(err.Error(), "source image filename") {
		t.Fatalf("setAltoImageFileName() error = %v, want missing filename error", err)
	}
}
