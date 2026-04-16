package main

import (
	"errors"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/krakenwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"golang.org/x/sync/errgroup"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	renderDPI          = 300
	diagramRegionType  = "GraphicZone-Diagram"
	pageNumberWidth    = 4
	defaultConcurrency = 4
	maxConcurrency     = 16
)

type pageJob struct {
	pdfBase     string
	pageNumber  int
	renderedPNG string
	outputDir   string
	model       string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) != 3 {
		return errors.New("usage: diagrams <input-dir> <output-dir> <segmentation-model>")
	}

	inputDir := strings.TrimSpace(args[0])
	outputDir := strings.TrimSpace(args[1])
	segmentationModel := strings.TrimSpace(args[2])

	if inputDir == "" {
		return errors.New("input dir is required")
	}
	if outputDir == "" {
		return errors.New("output dir is required")
	}
	if segmentationModel == "" {
		return errors.New("segmentation model is required")
	}

	info, err := os.Stat(inputDir)
	if err != nil {
		return fmt.Errorf("stat input dir %q: %w", inputDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("input path %q is not a directory", inputDir)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir %q: %w", outputDir, err)
	}

	pdfPaths, err := listPDFs(inputDir)
	if err != nil {
		return err
	}
	if len(pdfPaths) == 0 {
		return fmt.Errorf("input dir %q has no .pdf files", inputDir)
	}

	fmt.Printf("Found %d PDF files in %q, processing with segmentation model %q...\n", len(pdfPaths), inputDir, segmentationModel)
	for _, pdfPath := range pdfPaths {
		if err := processPDF(pdfPath, outputDir, segmentationModel); err != nil {
			return err
		}
	}

	return nil
}

func processPDF(pdfPath, outputRootDir, segmentationModel string) error {
	pdfBase := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	pdfOutputDir := filepath.Join(outputRootDir, pdfBase)

	if info, err := os.Stat(pdfOutputDir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("pdf output path %q exists and is not a directory", pdfOutputDir)
		}
		log.Printf("Skipping %s: output dir already exists at %s", pdfPath, pdfOutputDir)
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat pdf output dir %q: %w", pdfOutputDir, err)
	}

	if err := os.MkdirAll(pdfOutputDir, 0o755); err != nil {
		return fmt.Errorf("create pdf output dir %q: %w", pdfOutputDir, err)
	}
	log.Printf("Processing PDF %s -> %s", pdfPath, pdfOutputDir)

	rawDir, err := futils.MkdirTemp("diagrams-pages")
	if err != nil {
		return fmt.Errorf("create temp page dir: %w", err)
	}
	defer os.RemoveAll(rawDir)

	if err := formatcov.PDF2PNGs(pdfPath, rawDir, renderDPI); err != nil {
		return fmt.Errorf("render pdf %q: %w", pdfPath, err)
	}

	pageFiles, err := listRenderedPages(rawDir)
	if err != nil {
		return err
	}
	log.Printf("Rendered %d pages for %s", len(pageFiles), pdfPath)

	grp := new(errgroup.Group)
	grp.SetLimit(pipelineConcurrency())

	for _, pageFile := range pageFiles {
		pageFile := pageFile
		pageNum, err := pagesparser.FileNameToPage(pageFile)
		if err != nil {
			return fmt.Errorf("parse page number from %q: %w", pageFile, err)
		}

		job := pageJob{
			pdfBase:     pdfBase,
			pageNumber:  pageNum,
			renderedPNG: filepath.Join(rawDir, pageFile),
			outputDir:   pdfOutputDir,
			model:       segmentationModel,
		}

		grp.Go(func() error {
			return processPage(job)
		})
	}

	if err := grp.Wait(); err != nil {
		return fmt.Errorf("process pdf %q: %w", pdfPath, err)
	}

	log.Printf("Finished PDF %s", pdfPath)

	return nil
}

func pipelineConcurrency() int {
	n := runtime.NumCPU() / 2
	if n < defaultConcurrency {
		n = defaultConcurrency
	}
	if n > maxConcurrency {
		n = maxConcurrency
	}
	return n
}

func processPage(job pageJob) error {
	log.Printf("Starting page %04d for %s", job.pageNumber, job.pdfBase)

	tmpDir, err := futils.MkdirTemp(fmt.Sprintf("diagrams-%s-%04d", job.pdfBase, job.pageNumber))
	if err != nil {
		return fmt.Errorf("create page temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	denoisedPath := filepath.Join(tmpDir, "denoised.png")
	denoisedDeskewedPath := filepath.Join(tmpDir, "denoised_deskewed.png")
	deskewedPath := filepath.Join(tmpDir, "deskewed.png")
	segInputDir := filepath.Join(tmpDir, "segment-in")
	segOutputDir := filepath.Join(tmpDir, "segment-out")

	if err := os.MkdirAll(segInputDir, 0o755); err != nil {
		return fmt.Errorf("create segmentation input dir: %w", err)
	}

	if err := formatcov.DenoisePNGFile(job.renderedPNG, denoisedPath); err != nil {
		return fmt.Errorf("page %d denoise: %w", job.pageNumber, err)
	}
	log.Printf("Page %04d %s: denoised", job.pageNumber, job.pdfBase)

	angle, err := formatcov.EstimateDeskewAnglePNG(denoisedPath)
	if err != nil {
		return fmt.Errorf("page %d estimate deskew angle: %w", job.pageNumber, err)
	}
	log.Printf("Page %04d %s: deskew angle %.3f", job.pageNumber, job.pdfBase, angle)

	if err := formatcov.DeskewPNGFileWithAngle(denoisedPath, denoisedDeskewedPath, angle); err != nil {
		return fmt.Errorf("page %d deskew denoised image: %w", job.pageNumber, err)
	}
	if err := formatcov.DeskewPNGFileWithAngle(job.renderedPNG, deskewedPath, angle); err != nil {
		return fmt.Errorf("page %d deskew raw image: %w", job.pageNumber, err)
	}
	log.Printf("Page %04d %s: deskewed variants ready", job.pageNumber, job.pdfBase)

	segInputName := filepath.Base(denoisedDeskewedPath)
	segInputPath := filepath.Join(segInputDir, segInputName)
	if err := futils.CopyFile(denoisedDeskewedPath, segInputPath); err != nil {
		return fmt.Errorf("page %d prepare segmentation input: %w", job.pageNumber, err)
	}

	errCh, err := krakenwrapper.Segment(segInputDir, segOutputDir, job.model, []string{segInputName})
	if err != nil {
		return fmt.Errorf("page %d segment: %w", job.pageNumber, err)
	}
	if err := <-errCh; err != nil {
		return fmt.Errorf("page %d segment: %w", job.pageNumber, err)
	}
	log.Printf("Page %04d %s: segmentation complete", job.pageNumber, job.pdfBase)

	altoPath := filepath.Join(segOutputDir, strings.TrimSuffix(segInputName, filepath.Ext(segInputName))+".xml")
	crops, err := loadDiagramCrops(altoPath, deskewedPath)
	if err != nil {
		return fmt.Errorf("page %d load diagram regions: %w", job.pageNumber, err)
	}
	if len(crops) == 0 {
		log.Printf("Page %04d %s: no %s regions", job.pageNumber, job.pdfBase, diagramRegionType)
		return nil
	}

	if err := writeCrops(deskewedPath, crops, job.outputDir, job.pageNumber, false); err != nil {
		return fmt.Errorf("page %d write deskewed crops: %w", job.pageNumber, err)
	}
	if err := writeCrops(denoisedDeskewedPath, crops, job.outputDir, job.pageNumber, true); err != nil {
		return fmt.Errorf("page %d write denoised crops: %w", job.pageNumber, err)
	}
	log.Printf("Page %04d %s: wrote %d diagram crops for each variant", job.pageNumber, job.pdfBase, len(crops))

	return nil
}

func listPDFs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read input dir %q: %w", dir, err)
	}

	var pdfs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".pdf") {
			pdfs = append(pdfs, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(pdfs)
	return pdfs, nil
}

func listRenderedPages(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read rendered page dir %q: %w", dir, err)
	}

	var pages []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			pages = append(pages, entry.Name())
		}
	}

	sort.Strings(pages)
	return pages, nil
}

func loadDiagramCrops(altoPath, imagePath string) ([]alto.CropRegion, error) {
	doc, err := alto.LoadFromFile(altoPath)
	if err != nil {
		return nil, fmt.Errorf("load ALTO %q: %w", altoPath, err)
	}

	cfg, err := futils.ReadImageConfig(imagePath)
	if err != nil {
		return nil, fmt.Errorf("read image config %q: %w", imagePath, err)
	}

	bounds := image.Rect(0, 0, cfg.Width, cfg.Height)
	regions, err := alto.ExtractCropRegionsByCategory(doc, diagramRegionType, bounds)
	if err != nil {
		return nil, fmt.Errorf("extract ALTO crop regions: %w", err)
	}
	return regions, nil
}

func writeCrops(imagePath string, crops []alto.CropRegion, outputDir string, pageNumber int, denoised bool) error {
	src, err := futils.LoadImage(imagePath)
	if err != nil {
		return fmt.Errorf("decode image %q: %w", imagePath, err)
	}

	for _, crop := range crops {
		outName := cropFileName(pageNumber, crop.Index, denoised)
		outPath := filepath.Join(outputDir, outName)
		if err := futils.WritePNG(outPath, futils.CropImage(src, crop.Rect)); err != nil {
			return fmt.Errorf("write crop %q: %w", outPath, err)
		}
	}

	return nil
}

func cropFileName(pageNumber, diagramIndex int, denoised bool) string {
	base := fmt.Sprintf("%0*d_%d", pageNumberWidth, pageNumber, diagramIndex)
	if denoised {
		return base + "_denoised.png"
	}
	return base + ".png"
}
