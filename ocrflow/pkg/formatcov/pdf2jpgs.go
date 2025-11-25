package formatcov

import (
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
)

func PDF2JPGs(pdfPath, outDir string) error {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to open pdf: %w", err)
	}
	defer doc.Close()

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for i := 0; i < doc.NumPage(); i++ {
		img, err := doc.Image(i)
		if err != nil {
			return fmt.Errorf("failed to render page %d: %w", i+1, err)
		}

		outPath := filepath.Join(outDir, fmt.Sprintf("%04d.jpg", i+1))
		outFile, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create image file: %w", err)
		}

		if err := jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90}); err != nil {
			outFile.Close()
			return fmt.Errorf("failed to encode jpg for page %d: %w", i+1, err)
		}

		outFile.Close()
	}

	return nil
}
