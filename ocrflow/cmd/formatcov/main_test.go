package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestImageDirToPDFSupportsJPG(t *testing.T) {
	dir := t.TempDir()
	writeTestJPEG(t, filepath.Join(dir, "page-0002.jpg"), color.RGBA{R: 220, G: 30, B: 30, A: 255})
	writeTestJPEG(t, filepath.Join(dir, "page-0001.jpeg"), color.RGBA{R: 30, G: 80, B: 220, A: 255})
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}

	pdfPath := filepath.Join(t.TempDir(), "out.pdf")
	if err := imageDirToPDF(dir, pdfPath); err != nil {
		t.Fatalf("imageDirToPDF() error = %v", err)
	}

	got, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("read generated PDF: %v", err)
	}
	if !bytes.HasPrefix(got, []byte("%PDF-1.4\n")) {
		t.Fatalf("generated PDF does not have a PDF header")
	}
	if pageCount := bytes.Count(got, []byte("/Type /Page /Parent")); pageCount != 2 {
		t.Fatalf("generated PDF page count = %d, want 2", pageCount)
	}
	if !bytes.Contains(got, []byte("/Subtype /Image")) {
		t.Fatalf("generated PDF does not contain an image object")
	}
}

func writeTestJPEG(t *testing.T, path string, c color.Color) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 8, 6))
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, c)
		}
	}

	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test JPEG: %v", err)
	}
	defer out.Close()

	if err := jpeg.Encode(out, img, nil); err != nil {
		t.Fatalf("encode test JPEG: %v", err)
	}
}
