package formatcov

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
)

func PDF2PNGs(pdfPath, outDir string, dpi float64) error {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	n := doc.NumPage()
	if n == 0 {
		return fmt.Errorf("PDF has no pages")
	}

	for i := 0; i < n; i++ {
		img, err := doc.ImageDPI(i, dpi)
		if err != nil {
			return fmt.Errorf("failed to render page %d: %w", i+1, err)
		}

		filename := pagesparser.PageToPNGFilename(i + 1)
		outPath := filepath.Join(outDir, filename)

		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create file %q: %w", outPath, err)
		}

		if err := png.Encode(f, img); err != nil {
			_ = f.Close()
			return fmt.Errorf("failed to encode PNG for page %d: %w", i+1, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close file %q: %w", outPath, err)
		}

		log.Printf("Wrote [%d/%d]: %s", i+1, n, outPath)
	}

	return nil
}
