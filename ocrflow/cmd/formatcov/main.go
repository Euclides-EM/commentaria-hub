package main

import (
	"bytes"
	"compress/zlib"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/formatcov"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const supportedWorkflows = `Currently supported workflows:
  - PDF, PNG, JPEG, GIF, or a flat directory of PNG files -> denoised and deskewed PNG files
  - PAGE XML file or a flat directory of PAGE XML files -> ALTO XML files
  - flat directory of JPG, PNG, or GIF images -> PDF file`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags]\n\n%s\n\nFlags:\n", filepath.Base(os.Args[0]), supportedWorkflows)
		flag.PrintDefaults()
	}

	inputPath := flag.String("input", "", "input path: PDF, PNG/JPEG/GIF image, flat PNG directory, PAGE XML file, or flat PAGE XML directory")
	outputDir := flag.String("output-dir", "/tmp/formatcov", "output directory for processed PNG files or converted ALTO XML files")
	outputPDF := flag.String("output-pdf", "", "write a PDF from a flat image directory instead of processing images into -output-dir")
	pageRange := flag.String("range", "", "PDF pages to process (for example, 1,3-5); not supported for other input types")
	dpi := flag.Float64("dpi", 300, "resolution in DPI used when rendering PDF pages")
	flag.Parse()

	fmt.Println(supportedWorkflows)
	fmt.Println()

	if err := run(*inputPath, *outputDir, *outputPDF, *pageRange, *dpi); err != nil {
		log.Fatal(err)
	}
}

func run(inputPath, outputDir, outputPDF, pageRange string, dpi float64) error {
	inputPath = strings.TrimSpace(inputPath)
	outputDir = strings.TrimSpace(outputDir)
	outputPDF = strings.TrimSpace(outputPDF)
	pageRange = strings.TrimSpace(pageRange)

	if inputPath == "" {
		return errors.New("missing -input")
	}
	if outputDir == "" && outputPDF == "" {
		return errors.New("missing -output-dir")
	}
	if dpi <= 0 {
		return fmt.Errorf("dpi must be positive, got %v", dpi)
	}
	if outputPDF != "" {
		if pageRange != "" {
			return fmt.Errorf("page range is only supported for PDF input")
		}
		return imageDirToPDF(inputPath, outputPDF)
	}
	pageXML, err := formatcov.IsPageXMLInput(inputPath)
	if err != nil {
		return err
	}
	if pageXML {
		if pageRange != "" {
			return fmt.Errorf("page range is only supported for PDF input")
		}
		return formatcov.PageXMLFilesToALTO(inputPath, outputDir)
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

	denoiseDir, err := futils.MkdirTemp("formatcov-denoise")
	if err != nil {
		return fmt.Errorf("create denoise temp dir: %w", err)
	}
	defer os.RemoveAll(denoiseDir)

	if err := prepareInputPNGs(inputPath, rawDir, dpi, pages); err != nil {
		return err
	}
	if err := formatcov.DenoisePNGs(rawDir, denoiseDir); err != nil {
		return err
	}
	if err := formatcov.DeskewPNGs(denoiseDir, outputDir); err != nil {
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

	if strings.EqualFold(filepath.Ext(srcPath), ".png") {
		return copyFile(srcPath, dstPath)
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

func copyFile(srcPath, dstPath string) error {
	in, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy file: %w", err)
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
		if strings.EqualFold(filepath.Ext(entry.Name()), ".png") && !strings.HasSuffix(entry.Name(), ".snap.png") && !strings.HasSuffix(entry.Name(), ".stage.png") {
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
		fmt.Printf("preparing %q -> %q\n", srcPath, dstPath)
		if err := imageToPNG(srcPath, dstPath); err != nil {
			return fmt.Errorf("prepare %q: %w", srcPath, err)
		}
	}

	return nil
}

func imageDirToPDF(srcDir, pdfPath string) error {
	info, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("stat input %q: %w", srcDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("-output-pdf input must be a flat image directory")
	}

	imagePaths, err := listFlatImagePaths(srcDir)
	if err != nil {
		return err
	}
	if len(imagePaths) == 0 {
		return fmt.Errorf("input dir %q has no supported image files", srcDir)
	}

	if err := os.MkdirAll(filepath.Dir(pdfPath), 0o755); err != nil {
		return fmt.Errorf("create PDF output dir: %w", err)
	}

	var pages []pdfImagePage
	for _, imagePath := range imagePaths {
		fmt.Printf("adding %q to %q\n", imagePath, pdfPath)
		page, err := loadPDFImagePage(imagePath)
		if err != nil {
			return fmt.Errorf("load image %q: %w", imagePath, err)
		}
		pages = append(pages, page)
	}

	if err := writeImagePDF(pdfPath, pages); err != nil {
		return fmt.Errorf("write PDF %q: %w", pdfPath, err)
	}

	return nil
}

func listFlatImagePaths(srcDir string) ([]string, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, fmt.Errorf("read input dir %q: %w", srcDir, err)
	}

	var imagePaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif":
			imagePaths = append(imagePaths, filepath.Join(srcDir, entry.Name()))
		}
	}

	sort.Strings(imagePaths)
	return imagePaths, nil
}

type pdfImagePage struct {
	width  int
	height int
	data   []byte
}

func loadPDFImagePage(imagePath string) (pdfImagePage, error) {
	in, err := os.Open(imagePath)
	if err != nil {
		return pdfImagePage{}, fmt.Errorf("open image: %w", err)
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return pdfImagePage{}, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return pdfImagePage{}, fmt.Errorf("invalid image dimensions %dx%d", width, height)
	}

	rawRGB := make([]byte, 0, width*height*3)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rawRGB = append(rawRGB, byte(r>>8), byte(g>>8), byte(b>>8))
		}
	}

	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(rawRGB); err != nil {
		return pdfImagePage{}, fmt.Errorf("compress image data: %w", err)
	}
	if err := zw.Close(); err != nil {
		return pdfImagePage{}, fmt.Errorf("finish image compression: %w", err)
	}

	return pdfImagePage{
		width:  width,
		height: height,
		data:   compressed.Bytes(),
	}, nil
}

func writeImagePDF(pdfPath string, pages []pdfImagePage) error {
	out, err := os.Create(pdfPath)
	if err != nil {
		return fmt.Errorf("create PDF: %w", err)
	}
	defer out.Close()

	objectCount := 2 + len(pages)*3
	offsets := make([]int64, objectCount+1)

	write := func(format string, args ...any) error {
		_, err := fmt.Fprintf(out, format, args...)
		return err
	}
	writeObject := func(id int, body []byte) error {
		pos, err := out.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		offsets[id] = pos
		if err := write("%d 0 obj\n", id); err != nil {
			return err
		}
		if _, err := out.Write(body); err != nil {
			return err
		}
		return write("\nendobj\n")
	}

	if err := write("%%PDF-1.4\n"); err != nil {
		return err
	}
	if err := writeObject(1, []byte("<< /Type /Catalog /Pages 2 0 R >>")); err != nil {
		return err
	}

	var kids strings.Builder
	for i := range pages {
		fmt.Fprintf(&kids, "%d 0 R ", pageObjectID(i))
	}
	pagesBody := []byte(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", kids.String(), len(pages)))
	if err := writeObject(2, pagesBody); err != nil {
		return err
	}

	for i, page := range pages {
		pageID := pageObjectID(i)
		imageID := imageObjectID(i)
		contentID := contentObjectID(i)

		pageBody := []byte(fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %d %d] /Resources << /XObject << /Im%d %d 0 R >> >> /Contents %d 0 R >>",
			page.width, page.height, i+1, imageID, contentID,
		))
		if err := writeObject(pageID, pageBody); err != nil {
			return err
		}

		imageHeader := fmt.Sprintf(
			"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /FlateDecode /Length %d >>\nstream\n",
			page.width, page.height, len(page.data),
		)
		imageBody := append([]byte(imageHeader), page.data...)
		imageBody = append(imageBody, []byte("\nendstream")...)
		if err := writeObject(imageID, imageBody); err != nil {
			return err
		}

		content := []byte(fmt.Sprintf("q\n%d 0 0 %d 0 0 cm\n/Im%d Do\nQ\n", page.width, page.height, i+1))
		contentHeader := fmt.Sprintf("<< /Length %d >>\nstream\n", len(content))
		contentBody := append([]byte(contentHeader), content...)
		contentBody = append(contentBody, []byte("endstream")...)
		if err := writeObject(contentID, contentBody); err != nil {
			return err
		}
	}

	xrefOffset, err := out.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if err := write("xref\n0 %d\n0000000000 65535 f \n", objectCount+1); err != nil {
		return err
	}
	for id := 1; id <= objectCount; id++ {
		if err := write("%010d 00000 n \n", offsets[id]); err != nil {
			return err
		}
	}
	if err := write("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", objectCount+1, xrefOffset); err != nil {
		return err
	}

	return nil
}

func pageObjectID(pageIndex int) int {
	return 3 + pageIndex*3
}

func imageObjectID(pageIndex int) int {
	return pageObjectID(pageIndex) + 1
}

func contentObjectID(pageIndex int) int {
	return pageObjectID(pageIndex) + 2
}

func outputPNGName(inputPath string) string {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext) + ".png"
}
