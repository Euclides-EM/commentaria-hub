package formatcov

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
	"github.com/gen2brain/go-fitz"
	"github.com/samber/lo"
)

// PDF2PNGs renders all PDF pages to PNGs in outDir (page-0001.png, page-0002.png, ...).
func PDF2PNGs(pdfPath, outDir string, dpi float64) error {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	return pdf2pngsPages(doc, outDir, dpi, lo.RangeFrom(1, doc.NumPage()))
}

// PDF2PNGsWithPages renders only the given 1-based page numbers to PNGs in outDir.
// Page numbers must be within [1, doc.NumPage()].
func PDF2PNGsWithPages(pdfPath, outDir string, dpi float64, pages []int) error {
	if len(pages) == 0 {
		return fmt.Errorf("pages list is empty")
	}
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	n := doc.NumPage()
	if outOfRange := lo.Filter(pages, func(item int, _ int) bool {
		return item < 1 || item > n
	}); len(outOfRange) > 0 {
		return fmt.Errorf("pages out of range: %v (PDF has %d pages)", outOfRange, n)
	}
	return pdf2pngsPages(doc, outDir, dpi, pages)
}

func pdf2pngsPages(doc *fitz.Document, outDir string, dpi float64, pages []int) error {
	n := doc.NumPage()
	if n == 0 {
		return fmt.Errorf("PDF has no pages")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	for i, pageNum := range pages {
		img, err := doc.ImageDPI(pageNum-1, dpi)
		if err != nil {
			return fmt.Errorf("failed to render page %d: %w", pageNum, err)
		}

		filename := pagesparser.PageToPNGFilename(pageNum)
		outPath := filepath.Join(outDir, filename)

		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create file %q: %w", outPath, err)
		}

		if err := png.Encode(f, img); err != nil {
			_ = f.Close()
			return fmt.Errorf("failed to encode PNG for page %d: %w", pageNum, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close file %q: %w", outPath, err)
		}

		log.Printf("Wrote [%d/%d]: %s (page %d)", i+1, len(pages), outPath, pageNum)
	}

	return nil
}
