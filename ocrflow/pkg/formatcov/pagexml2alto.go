package formatcov

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
)

const altoNamespace = "http://www.loc.gov/standards/alto/ns-v4#"

type pageXMLDocument struct {
	XMLName xml.Name    `xml:"PcGts"`
	Page    pageXMLPage `xml:"Page"`
}

type pageXMLPage struct {
	ImageFilename string          `xml:"imageFilename,attr"`
	ImageWidth    int             `xml:"imageWidth,attr"`
	ImageHeight   int             `xml:"imageHeight,attr"`
	TextRegions   []pageXMLRegion `xml:"TextRegion"`
}

type pageXMLRegion struct {
	ID     string        `xml:"id,attr"`
	Coords pageXMLCoords `xml:"Coords"`
	Lines  []pageXMLLine `xml:"TextLine"`
}

type pageXMLLine struct {
	ID        string             `xml:"id,attr"`
	Coords    pageXMLCoords      `xml:"Coords"`
	Baseline  pageXMLCoords      `xml:"Baseline"`
	TextEquiv []pageXMLTextEquiv `xml:"TextEquiv"`
	Words     []pageXMLWord      `xml:"Word"`
}

type pageXMLWord struct {
	Coords    pageXMLCoords      `xml:"Coords"`
	TextEquiv []pageXMLTextEquiv `xml:"TextEquiv"`
}

type pageXMLCoords struct {
	Points string `xml:"points,attr"`
}

type pageXMLTextEquiv struct {
	Index   int    `xml:"index,attr"`
	Unicode string `xml:"Unicode"`
}

// IsPageXMLInput reports whether inputPath is an XML file or a flat directory
// containing XML files. Mixed XML/PNG directories are rejected because their
// intended conversion mode is ambiguous.
func IsPageXMLInput(inputPath string) (bool, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return false, fmt.Errorf("stat input %q: %w", inputPath, err)
	}
	if !info.IsDir() {
		return strings.EqualFold(filepath.Ext(inputPath), ".xml"), nil
	}
	entries, err := os.ReadDir(inputPath)
	if err != nil {
		return false, fmt.Errorf("read input dir %q: %w", inputPath, err)
	}
	var hasXML, hasPNG bool
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(entry.Name())) {
		case ".xml":
			hasXML = true
		case ".png":
			hasPNG = true
		}
	}
	if hasXML && hasPNG {
		return false, fmt.Errorf("input dir %q contains both PAGE XML and PNG files; use a directory containing only one input format", inputPath)
	}
	return hasXML, nil
}

// PageXML2ALTO converts one PAGE XML document into ALTO 4. PAGE regions and
// lines retain their IDs, polygons, and baselines. If PAGE has no word-level
// segmentation, the complete line transcription is emitted as one ALTO String.
func PageXML2ALTO(data []byte, pageID string) (*alto.Alto, error) {
	var source pageXMLDocument
	if err := xml.Unmarshal(data, &source); err != nil {
		return nil, fmt.Errorf("decode PAGE XML: %w", err)
	}
	if source.XMLName.Local != "PcGts" {
		return nil, fmt.Errorf("decode PAGE XML: expected PcGts root element, got %q", source.XMLName.Local)
	}
	if source.Page.ImageWidth <= 0 || source.Page.ImageHeight <= 0 {
		return nil, fmt.Errorf("decode PAGE XML: invalid page dimensions %dx%d", source.Page.ImageWidth, source.Page.ImageHeight)
	}
	if pageID == "" {
		pageID = "page"
	}

	blocks := make([]alto.TextBlock, 0, len(source.Page.TextRegions))
	for regionIndex, region := range source.Page.TextRegions {
		block, err := pageRegionToALTO(region, regionIndex)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return &alto.Alto{
		Xmlns:          altoNamespace,
		XmlnsXsi:       "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: altoNamespace + " http://www.loc.gov/standards/alto/v4/alto-4-3.xsd",
		Description: alto.Description{
			MeasurementUnit: "pixel",
			SourceImageInformation: alto.SourceImageInformation{
				FileName: source.Page.ImageFilename,
			},
		},
		Layout: alto.Layout{Page: []alto.Page{{
			Width:  source.Page.ImageWidth,
			Height: source.Page.ImageHeight,
			ID:     pageID,
			PrintSpace: alto.PrintSpace{
				HPOS:       0,
				VPOS:       0,
				Width:      float64(source.Page.ImageWidth),
				Height:     float64(source.Page.ImageHeight),
				TextBlocks: blocks,
			},
		}}},
	}, nil
}

// PageXMLFilesToALTO converts a PAGE XML file or every .xml file in a flat
// directory. Output filenames match their input filenames.
func PageXMLFilesToALTO(inputPath, outputDir string) error {
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("stat PAGE XML input %q: %w", inputPath, err)
	}

	var paths []string
	if info.IsDir() {
		entries, err := os.ReadDir(inputPath)
		if err != nil {
			return fmt.Errorf("read PAGE XML directory %q: %w", inputPath, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".xml") {
				paths = append(paths, filepath.Join(inputPath, entry.Name()))
			}
		}
		sort.Strings(paths)
	} else if strings.EqualFold(filepath.Ext(inputPath), ".xml") {
		paths = []string{inputPath}
	} else {
		return fmt.Errorf("PAGE XML input %q must be an .xml file or directory", inputPath)
	}
	if len(paths) == 0 {
		return fmt.Errorf("PAGE XML directory %q has no .xml files", inputPath)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create ALTO output directory %q: %w", outputDir, err)
	}

	for _, inputFile := range paths {
		base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
		outputFile := filepath.Join(outputDir, base+".xml")
		same, err := sameFilePath(inputFile, outputFile)
		if err != nil {
			return err
		}
		if same {
			return fmt.Errorf("refusing to overwrite PAGE XML input %q; choose a different output directory", inputFile)
		}
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("read PAGE XML %q: %w", inputFile, err)
		}
		doc, err := PageXML2ALTO(data, base)
		if err != nil {
			return fmt.Errorf("convert PAGE XML %q: %w", inputFile, err)
		}
		fmt.Printf("converting PAGE XML %q -> %q\n", inputFile, outputFile)
		if err := alto.SaveToFile(doc, outputFile); err != nil {
			return fmt.Errorf("write ALTO %q: %w", outputFile, err)
		}
	}
	return nil
}

func sameFilePath(a, b string) (bool, error) {
	infoA, err := os.Stat(a)
	if err != nil {
		return false, fmt.Errorf("stat input path %q: %w", a, err)
	}
	if infoB, err := os.Stat(b); err == nil && os.SameFile(infoA, infoB) {
		return true, nil
	} else if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("stat output path %q: %w", b, err)
	}
	absA, err := filepath.Abs(a)
	if err != nil {
		return false, fmt.Errorf("resolve input path %q: %w", a, err)
	}
	absB, err := filepath.Abs(b)
	if err != nil {
		return false, fmt.Errorf("resolve output path %q: %w", b, err)
	}
	return filepath.Clean(absA) == filepath.Clean(absB), nil
}

func pageRegionToALTO(region pageXMLRegion, regionIndex int) (alto.TextBlock, error) {
	points, box, err := pagePointsAndBounds(region.Coords.Points)
	if err != nil {
		return alto.TextBlock{}, fmt.Errorf("region %q coordinates: %w", region.ID, err)
	}
	id := region.ID
	if id == "" {
		id = fmt.Sprintf("region_%d", regionIndex+1)
	}
	lines := make([]alto.TextLine, 0, len(region.Lines))
	for lineIndex, line := range region.Lines {
		converted, err := pageLineToALTO(line, id, lineIndex)
		if err != nil {
			return alto.TextBlock{}, err
		}
		lines = append(lines, converted)
	}
	return alto.TextBlock{
		ID: id, HPOS: box.minX, VPOS: box.minY, Width: box.width(), Height: box.height(),
		Shape: alto.Shape{Polygon: alto.Polygon{Points: points}}, Lines: lines,
	}, nil
}

func pageLineToALTO(line pageXMLLine, regionID string, lineIndex int) (alto.TextLine, error) {
	points, box, err := pagePointsAndBounds(line.Coords.Points)
	if err != nil {
		return alto.TextLine{}, fmt.Errorf("line %q in region %q coordinates: %w", line.ID, regionID, err)
	}
	id := line.ID
	if id == "" {
		id = fmt.Sprintf("%s_line_%d", regionID, lineIndex+1)
	}
	result := alto.TextLine{
		ID: id, HPOS: box.minX, VPOS: box.minY, Width: box.width(), Height: box.height(),
		Shape: alto.Shape{Polygon: alto.Polygon{Points: points}},
	}
	if strings.TrimSpace(line.Baseline.Points) != "" {
		baseline, _, err := pagePointsAndBounds(line.Baseline.Points)
		if err != nil {
			return alto.TextLine{}, fmt.Errorf("line %q baseline: %w", id, err)
		}
		result.Baseline = baseline
	}

	if len(line.Words) == 0 {
		if text := primaryText(line.TextEquiv); text != "" {
			result.Strings = []alto.String{{
				Content: text, HPOS: box.minX, VPOS: box.minY, Width: box.width(), Height: box.height(),
			}}
		}
		return result, nil
	}
	for wordIndex, word := range line.Words {
		_, wordBox, err := pagePointsAndBounds(word.Coords.Points)
		if err != nil {
			return alto.TextLine{}, fmt.Errorf("word %d in line %q coordinates: %w", wordIndex+1, id, err)
		}
		if text := primaryText(word.TextEquiv); text != "" {
			result.Strings = append(result.Strings, alto.String{
				Content: text, HPOS: wordBox.minX, VPOS: wordBox.minY, Width: wordBox.width(), Height: wordBox.height(),
			})
		}
	}
	return result, nil
}

func primaryText(equivs []pageXMLTextEquiv) string {
	if len(equivs) == 0 {
		return ""
	}
	best := 0
	for i := 1; i < len(equivs); i++ {
		if equivs[i].Index < equivs[best].Index {
			best = i
		}
	}
	return strings.TrimSpace(equivs[best].Unicode)
}

type pageBounds struct{ minX, minY, maxX, maxY float64 }

func (b pageBounds) width() float64  { return b.maxX - b.minX }
func (b pageBounds) height() float64 { return b.maxY - b.minY }

func pagePointsAndBounds(raw string) (string, pageBounds, error) {
	tokens := strings.Fields(raw)
	if len(tokens) < 2 {
		return "", pageBounds{}, fmt.Errorf("missing points")
	}
	values := make([]float64, 0, len(tokens)*2)
	for _, token := range tokens {
		xy := strings.Split(token, ",")
		if len(xy) != 2 {
			return "", pageBounds{}, fmt.Errorf("invalid point %q", token)
		}
		for _, component := range xy {
			value, err := strconv.ParseFloat(component, 64)
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
				return "", pageBounds{}, fmt.Errorf("invalid coordinate %q", component)
			}
			values = append(values, value)
		}
	}
	b := pageBounds{minX: values[0], maxX: values[0], minY: values[1], maxY: values[1]}
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.FormatFloat(value, 'f', -1, 64)
		if i%2 == 0 {
			b.minX, b.maxX = math.Min(b.minX, value), math.Max(b.maxX, value)
		} else {
			b.minY, b.maxY = math.Min(b.minY, value), math.Max(b.maxY, value)
		}
	}
	return strings.Join(parts, " "), b, nil
}
