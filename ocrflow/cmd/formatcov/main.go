package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func main() {
	inputPath := flag.String("input", "", "path to input image, PNG directory, or PDF")
	outputDir := flag.String("output-dir", "/tmp/formatcov", "directory for processed PNG output")
	pageRange := flag.String("range", "", "optional PDF page range, e.g. 1,3-5")
	dpi := flag.Float64("dpi", 300, "PDF render DPI")
	flag.Parse()

	if err := run(*inputPath, *outputDir, *pageRange, *dpi); err != nil {
		log.Fatal(err)
	}
}

func run(inputPath, outputDir, pageRange string, dpi float64) error {
	inputPath = strings.TrimSpace(inputPath)
	outputDir = strings.TrimSpace(outputDir)
	pageRange = strings.TrimSpace(pageRange)

	if inputPath == "" {
		return errors.New("missing -input")
	}
	if outputDir == "" {
		return errors.New("missing -output-dir")
	}
	if dpi <= 0 {
		return fmt.Errorf("dpi must be positive, got %v", dpi)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir %q: %w", outputDir, err)
	}

	var pages []int
	if pageRange != "" {
		var err error
		pages, err = pagesparser.IntRange(pageRange)
		if err != nil {
			return fmt.Errorf("invalid range %q: %w", pageRange, err)
		}
	}

	rawDir, err := futils.MkdirTemp("formatcov-raw")
	if err != nil {
		return fmt.Errorf("create raw temp dir: %w", err)
	}
	defer os.RemoveAll(rawDir)

	deskewDir, err := futils.MkdirTemp("formatcov-deskew")
	if err != nil {
		return fmt.Errorf("create deskew temp dir: %w", err)
	}
	defer os.RemoveAll(deskewDir)

	if err := prepareInputPNGs(inputPath, rawDir, dpi, pages); err != nil {
		return err
	}
	if err := formatcov.DeskewPNGs(rawDir, deskewDir); err != nil {
		return err
	}
	if err := formatcov.DenoisePNGs(deskewDir, outputDir); err != nil {
		return err
	}

	return nil
}

func prepareInputPNGs(inputPath, outDir string, dpi float64, pages []int) error {
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("stat input %q: %w", inputPath, err)
	}
	if info.IsDir() {
		if len(pages) > 0 {
			return fmt.Errorf("page range is only supported for PDF input")
		}
		return pngDirToPNGs(inputPath, outDir)
	}

	ext := strings.ToLower(filepath.Ext(inputPath))
	switch ext {
	case ".pdf":
		if len(pages) > 0 {
			return formatcov.PDF2PNGsWithPages(inputPath, outDir, dpi, pages)
		}
		return formatcov.PDF2PNGs(inputPath, outDir, dpi)
	default:
		if len(pages) > 0 {
			return fmt.Errorf("page range is only supported for PDF input")
		}
		return imageToPNG(inputPath, filepath.Join(outDir, outputPNGName(inputPath)))
	}
}

func imageToPNG(srcPath, dstPath string) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	in, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	defer out.Close()

	if err := png.Encode(out, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	return nil
}

func pngDirToPNGs(srcDir, outDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read input dir %q: %w", srcDir, err)
	}

	var pngNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			pngNames = append(pngNames, entry.Name())
		}
	}

	if len(pngNames) == 0 {
		return fmt.Errorf("input dir %q has no .png files", srcDir)
	}

	sort.Strings(pngNames)
	for _, name := range pngNames {
		srcPath := filepath.Join(srcDir, name)
		dstPath := filepath.Join(outDir, outputPNGName(name))
		if err := imageToPNG(srcPath, dstPath); err != nil {
			return fmt.Errorf("prepare %q: %w", srcPath, err)
		}
	}

	return nil
}

func outputPNGName(inputPath string) string {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext) + ".png"
}
