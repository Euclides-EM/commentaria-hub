package krakenwrapper

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/beevik/etree"
)

// RecognizeTextWithMapping overwrites existing ALTO files with OCR-ed ALTO output.
// Each pair contains an absolute image path followed by its absolute ALTO path.
func RecognizeTextWithMapping(imgAndAltoPaths [][2]string, ocrModel string) (<-chan error, error) {
	if strings.TrimSpace(ocrModel) == "" {
		return nil, fmt.Errorf("ocr model is required")
	}
	if err := validateImgAndAltoPaths(imgAndAltoPaths); err != nil {
		return nil, err
	}
	if len(imgAndAltoPaths) == 0 {
		ch := make(chan error, 1)
		close(ch)
		return ch, nil
	}
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- runPairsOCRUsingExistingAlto(imgAndAltoPaths, ocrModel)
	}()
	return errCh, nil
}

func validateImgAndAltoPaths(imgAndAltoPaths [][2]string) error {
	for _, imgAndAltoPath := range imgAndAltoPaths {
		imgPath := imgAndAltoPath[0]
		altoPath := imgAndAltoPath[1]

		// Validate image
		if !slices.Contains(imageFormats, strings.ToLower(filepath.Ext(imgPath))) {
			return fmt.Errorf("image input file %s is not a supported image format (TIFF/PNG)", imgPath)
		}
		if !filepath.IsAbs(imgPath) {
			return fmt.Errorf("image input file %s is not an absolute path", imgPath)
		}
		if _, err := os.Stat(imgPath); err != nil {
			return fmt.Errorf("image input file %s does not exist: %w", imgPath, err)
		}

		// Validate ALTO
		if strings.ToLower(filepath.Ext(altoPath)) != ".xml" {
			return fmt.Errorf("ALTO file %s does not have .xml extension", altoPath)
		}
		if !filepath.IsAbs(altoPath) {
			return fmt.Errorf("ALTO file %s is not an absolute path", altoPath)
		}
		if _, err := os.Stat(altoPath); err != nil {
			return fmt.Errorf("ALTO file %s does not exist: %w", altoPath, err)
		}
	}
	return nil
}

// Uses existing ALTO (segmentation+lines) as input and overwrites it with OCR-ed ALTO.
func runPairsOCRUsingExistingAlto(imgAndAltoPaths [][2]string, ocrModel string) error {

	// create temp dir that includes the ato and images
	tmpDir, err := futils.MkdirTemp("kraken_ocr_reuse_alto")
	if err != nil {
		return fmt.Errorf("could not create temp dir for kraken OCR reuse ALTO: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	tmpToOrigAltoPath := make(map[string]string) // tmp ALTO path -> original ALTO path
	imgAndAltoTmpPaths := make([][2]string, 0)
	for _, pair := range imgAndAltoPaths {
		imgPath := pair[0]
		altoPath := pair[1]

		tmpImgPath := filepath.Join(tmpDir, filepath.Base(imgPath))
		tmpAltoPath := filepath.Join(tmpDir, filepath.Base(altoPath))

		if err := futils.CopyFile(imgPath, tmpImgPath); err != nil {
			return fmt.Errorf("could not link image %s to temp location: %w", imgPath, err)
		}
		if err := futils.CopyFile(altoPath, tmpAltoPath); err != nil {
			return fmt.Errorf("could not copy ALTO %s to temp location: %w", altoPath, err)
		}
		hasLines, err := prepareAltoForOCR(tmpAltoPath, filepath.Base(tmpImgPath))
		if err != nil {
			return fmt.Errorf("could not prepare ALTO %s for Kraken OCR: %w", altoPath, err)
		}
		if !hasLines {
			continue
		}

		imgAndAltoTmpPaths = append(imgAndAltoTmpPaths, [2]string{tmpImgPath, tmpAltoPath})
		tmpToOrigAltoPath[tmpAltoPath] = altoPath
	}

	// Run Kraken OCR in parallel on pairs of (img, alto) files. Kraken will read the ALTO for segmentation

	workers := runtime.NumCPU()
	if workers > maxParallelKraken {
		workers = maxParallelKraken
	}
	if workers > len(imgAndAltoTmpPaths) {
		workers = len(imgAndAltoTmpPaths)
	}
	if workers < 1 {
		workers = 1
	}
	chunks := chunkPairs(imgAndAltoTmpPaths, workers)

	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	wg.Add(len(chunks))
	for _, chunk := range chunks {
		chunk := chunk
		go func() {
			defer wg.Done()
			if err := runKrakenOCRReuseAlto(chunk, ocrModel); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}

	// Validate all outputs before replacing any original ALTO.
	for _, pair := range imgAndAltoTmpPaths {
		altoTmpPath := pair[1]
		tmp := tmpOcredPath(altoTmpPath)
		info, err := os.Stat(tmp)
		if err != nil {
			return fmt.Errorf("expected OCR output file %s does not exist: %w", tmp, err)
		}
		if info.Size() == 0 {
			return fmt.Errorf("Kraken produced an empty OCR output file %s", tmp)
		}
		if err := RemovePathFromAltoImgFileName(tmp, tmp); err != nil {
			return fmt.Errorf("could not normalize source image name in OCR output %s: %w", tmp, err)
		}
	}
	for _, pair := range imgAndAltoTmpPaths {
		altoTmpPath := pair[1]
		originalPath := tmpToOrigAltoPath[altoTmpPath]
		if err := replaceFile(tmpOcredPath(altoTmpPath), originalPath); err != nil {
			return fmt.Errorf("could not replace ALTO %s with OCR output: %w", originalPath, err)
		}
	}
	return nil
}

func replaceFile(src, dst string) error {
	replacement, err := futils.CreateTempInDir(filepath.Dir(dst), "ocr-replace-*.xml")
	if err != nil {
		return err
	}
	replacementPath := replacement.Name()
	if err := replacement.Close(); err != nil {
		_ = os.Remove(replacementPath)
		return err
	}
	defer os.Remove(replacementPath)
	if err := futils.CopyFile(src, replacementPath); err != nil {
		return err
	}
	return os.Rename(replacementPath, dst)
}

// prepareAltoForOCR makes the small set of changes Kraken 5.3 needs when an
// existing ALTO document is used as recognition input. Unknown ALTO elements
// and attributes are retained.
func prepareAltoForOCR(altoPath, imageName string) (bool, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(altoPath); err != nil {
		return false, err
	}
	lines := doc.FindElements("//TextLine")
	if len(lines) == 0 {
		return false, nil
	}

	fileName := doc.FindElement("//fileName")
	if fileName == nil {
		return false, fmt.Errorf("ALTO has no sourceImageInformation/fileName element")
	}
	fileName.SetText(imageName)

	// Kraken's ALTO serializer cannot handle empty region placeholders that
	// have neither geometry nor lines. They carry no annotation data.
	for _, block := range doc.FindElements("//TextBlock") {
		if hasValidBoundary(block) || validBoundingBox(block) || len(block.SelectElements("TextLine")) > 0 {
			continue
		}
		block.Parent().RemoveChild(block)
	}

	for _, line := range lines {
		if !validBaselinePoints(line.SelectAttrValue("BASELINE", "")) {
			return false, fmt.Errorf("text line %q has no valid baseline", line.SelectAttrValue("ID", ""))
		}

		// A line TAGREFS makes Kraken enter multi-model recognition. This rule
		// applies one model to every line, so do not expose line tags as model
		// selectors in the staged input. Region tags remain unchanged.
		line.RemoveAttr("TAGREFS")

		shape := line.SelectElement("Shape")
		var polygon *etree.Element
		if shape != nil {
			polygon = shape.SelectElement("Polygon")
		}
		if polygon != nil && validPolygonPoints(polygon.SelectAttrValue("POINTS", "")) {
			continue
		}

		points := line.SelectAttrValue("POINTS", "")
		if !validPolygonPoints(points) {
			x, y, width, height, ok := boundingBox(line)
			if !ok {
				return false, fmt.Errorf("text line %q has no valid polygon or bounding box", line.SelectAttrValue("ID", ""))
			}
			points = fmt.Sprintf("%g %g %g %g %g %g %g %g %g %g",
				x, y, x+width, y, x+width, y+height, x, y+height, x, y)
		}
		if shape == nil {
			shape = line.CreateElement("Shape")
		}
		if polygon == nil {
			polygon = shape.CreateElement("Polygon")
		}
		polygon.RemoveAttr("POINTS")
		polygon.CreateAttr("POINTS", points)
	}

	doc.Indent(2)
	if err := doc.WriteToFile(altoPath); err != nil {
		return false, err
	}
	return len(lines) > 0, nil
}

func hasValidBoundary(element *etree.Element) bool {
	shape := element.SelectElement("Shape")
	if shape == nil {
		return false
	}
	polygon := shape.SelectElement("Polygon")
	return polygon != nil && validPolygonPoints(polygon.SelectAttrValue("POINTS", ""))
}

func validBoundingBox(element *etree.Element) bool {
	_, _, _, _, ok := boundingBox(element)
	return ok
}

func boundingBox(element *etree.Element) (x, y, width, height float64, ok bool) {
	values := []*float64{&x, &y, &width, &height}
	for i, name := range []string{"HPOS", "VPOS", "WIDTH", "HEIGHT"} {
		value, err := strconv.ParseFloat(element.SelectAttrValue(name, ""), 64)
		if err != nil {
			return 0, 0, 0, 0, false
		}
		*values[i] = value
	}
	return x, y, width, height, width > 0 && height > 0
}

func validPolygonPoints(points string) bool {
	fields := strings.Fields(strings.NewReplacer(",", " ", "(", " ", ")", " ").Replace(points))
	if len(fields) < 6 || len(fields)%2 != 0 {
		return false
	}
	xs := make(map[float64]struct{})
	ys := make(map[float64]struct{})
	for i, field := range fields {
		value, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return false
		}
		if i%2 == 0 {
			xs[value] = struct{}{}
		} else {
			ys[value] = struct{}{}
		}
	}
	return len(xs) > 1 && len(ys) > 1
}

func validBaselinePoints(points string) bool {
	fields := strings.Fields(strings.NewReplacer(",", " ", "(", " ", ")", " ").Replace(points))
	if len(fields) < 4 || len(fields)%2 != 0 {
		return false
	}
	for _, field := range fields {
		if _, err := strconv.ParseFloat(field, 64); err != nil {
			return false
		}
	}
	return true
}

func runKrakenOCRReuseAlto(pairs [][2]string, ocrModel string) error {

	if len(pairs) == 0 {
		return nil
	}

	for _, p := range pairs {
		altoOutTmp := tmpOcredPath(p[1])

		// Ensure output dir exists
		if err := os.MkdirAll(filepath.Dir(altoOutTmp), 0755); err != nil {
			return fmt.Errorf("could not create ALTO output directory %s: %w", filepath.Dir(altoOutTmp), err)
		}
	}

	if err := envexec.PythonCmd("kraken", krakenOCRReuseAltoArgs(pairs, ocrModel)...); err != nil {
		return fmt.Errorf("kraken ocr (reuse alto) failed: %w", err)
	}

	return nil
}

func krakenOCRReuseAltoArgs(pairs [][2]string, ocrModel string) []string {
	args := []string{"--alto", "--format-type", "alto", "--raise-on-error"}
	args = append(args, krakenDeviceArgs()...)
	// Kraken reads both the existing segmentation and source image reference
	// from the ALTO input. pair[0] is the image staged beside pair[1].
	for _, pair := range pairs {
		args = append(args, "-i", pair[1], tmpOcredPath(pair[1]))
	}
	// Kraken 5.3 marks XML input as tag-aware even when it has no custom line
	// tags. Map its normalized "default" line type explicitly to this model.
	return append(args, "ocr", "-m", "default:"+ocrModel)
}

func tmpOcredPath(finalAltoPath string) string {
	return finalAltoPath + ".ocr.tmp"
}
