package formatcov

import (
	"bufio"
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var SubDirs = []string{"", "train", "valid", "test"}

func Yolo2Alto(src string, dst string) error {
	lm, err := loadLabelmapFromDataYml(path.Join(src, "data.yaml"))
	if err != nil {
		log.Printf("WARNING: load labelmap from data.yaml failed, this may be ok, trying labelmap.txt: %v", err)
	} else {
		fmt.Println("Loaded labelmap from data.yaml with", len(lm), "labels")
	}
	for _, subDir := range SubDirs {
		if _, err := os.Stat(filepath.Join(src, subDir, "labels")); os.IsNotExist(err) {
			continue
		}
		log.Println("Found yolo data in", filepath.Join(src, subDir), ", converting to ALTO in", dst)
		if err := yolo2Alto(filepath.Join(src, subDir), dst, lm); err != nil {
			return err
		}
	}
	return nil
}

//go:embed templates/alto_template.xml
var altoTemplate string

func yolo2Alto(src string, dst string, labelmap []string) error {
	if len(labelmap) == 0 {
		var err error
		labelmap, err = loadLabelmapFromTXTFile(src)
		if err != nil {
			return err
		}
	}
	if len(labelmap) == 0 {
		return fmt.Errorf("no labels found in labelmap.txt or data.yaml under")
	}

	otherTags := buildOtherTags(labelmap)

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create dst dir %s: %w", dst, err)
	}

	pattern := filepath.Join(src, "labels", "*.txt")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob %s: %w", pattern, err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no .txt files found in %s", src)
	}

	for _, file := range files {
		// skip labelmap.txt itself if it matches the glob
		if filepath.Base(file) == "labelmap.txt" {
			continue
		}
		if err := convertSingleYoloFile(file, dst, otherTags); err != nil {
			return err
		}
	}

	return nil
}

// --- labelmap loading ---

func loadLabelmapFromTXTFile(src string) ([]string, error) {
	lmPath := filepath.Join(src, "labelmap.txt")
	if !fileExists(lmPath) {
		return nil, fmt.Errorf("labelmap.txt not found in %s", src)
	}
	labels, err := readLabelmapTxt(lmPath)
	if err != nil {
		return nil, fmt.Errorf("read labelmap.txt: %w", err)
	}
	if len(labels) > 0 {
		return labels, nil
	}
	return nil, fmt.Errorf("no labels found in labelmap.txt in %s", src)
}

func loadLabelmapFromDataYml(dataYML string) ([]string, error) {
	if !fileExists(dataYML) {
		return nil, fmt.Errorf("data.yaml not found in %s", dataYML)
	}
	labels, err := readLabelmapFromYAML(dataYML)
	if err != nil {
		return nil, fmt.Errorf("read data.yaml: %w", err)
	}
	if len(labels) > 0 {
		return labels, nil
	}
	return nil, fmt.Errorf("no labels found in in %s", dataYML)
}

func readLabelmapTxt(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var labels []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		labels = append(labels, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return labels, nil
}

type yoloDataConfig struct {
	Names []string `yaml:"names"`
}

func readLabelmapFromYAML(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg yoloDataConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// cfg.Names should be like:
	// names: ['CatchWordZone', 'DropCapitalZone', ...]
	return cfg.Names, nil
}

// --- ALTO building ---

func buildOtherTags(labelmap []string) string {
	var b strings.Builder
	for idx, zone := range labelmap {
		if idx > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(
			&b,
			`<OtherTag ID="BT%03d" LABEL="%s" DESCRIPTION="block type %s"/>`,
			idx,
			xmlEscape(zone),
			xmlEscape(zone),
		)
	}
	return b.String()
}

var pagePattern = regexp.MustCompile(`^(page-\d{4}).*`)

func CutPagePrefix(s string) (string, bool) {
	m := pagePattern.FindStringSubmatch(s)
	if len(m) > 1 {
		return m[1], true
	}
	return s, false
}

func convertSingleYoloFile(yoloPath string, dstDir string, otherTags string) error {
	base := strings.TrimSuffix(filepath.Base(yoloPath), filepath.Ext(yoloPath))

	var outputBase = base
	if p, ok := CutPagePrefix(base); ok {
		outputBase = p
	}

	imgFileName := base + ".jpg"

	dir := filepath.Dir(yoloPath)

	candidate1 := filepath.Join(dir, "..", "images", imgFileName)
	candidate2 := filepath.Join(dir, "../..", "images", imgFileName)

	var imgPath string

	if fileExists(candidate1) {
		imgPath = candidate1
	} else if fileExists(candidate2) {
		imgPath = candidate2
	} else {
		return fmt.Errorf("cannot find the image for %s (tried %s and %s)", imgFileName, candidate1, candidate2)
	}

	imgWidth, imgHeight, err := imageSize(imgPath)
	if err != nil {
		return fmt.Errorf("read image %s: %w", imgPath, err)
	}

	zones, err := buildTextBlocksFromYolo(yoloPath, imgWidth, imgHeight)
	if err != nil {
		return err
	}

	xmlContent := altoTemplate
	xmlContent = strings.ReplaceAll(xmlContent, "%Filename%", xmlEscape(outputBase+".png"))
	xmlContent = strings.ReplaceAll(xmlContent, "%Width%", strconv.Itoa(imgWidth))
	xmlContent = strings.ReplaceAll(xmlContent, "%Height%", strconv.Itoa(imgHeight))
	xmlContent = strings.ReplaceAll(xmlContent, "%Tags%", otherTags)
	xmlContent = strings.ReplaceAll(xmlContent, "%Textblocks%", strings.Join(zones, ""))

	xmlPath := filepath.Join(dstDir, outputBase+".xml")
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0o644); err != nil {
		return fmt.Errorf("write xml %s: %w", xmlPath, err)
	}

	return nil
}

func buildTextBlocksFromYolo(yoloPath string, imgWidth, imgHeight int) ([]string, error) {
	f, err := os.Open(yoloPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var zones []string
	scanner := bufio.NewScanner(f)
	lineIdx := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return nil, fmt.Errorf("invalid YOLO line in %s: %q", yoloPath, line)
		}

		classID, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, fmt.Errorf("parse class id in %s line %q: %w", yoloPath, line, err)
		}
		cx, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return nil, err
		}
		cy, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return nil, err
		}
		w, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			return nil, err
		}
		h, err := strconv.ParseFloat(fields[4], 64)
		if err != nil {
			return nil, err
		}

		// YOLO: class cx cy w h, normalized 0..1
		x0f := float64(imgWidth) * (cx - w/2.0)
		x1f := float64(imgWidth) * (cx + w/2.0)
		y0f := float64(imgHeight) * (cy - h/2.0)
		y1f := float64(imgHeight) * (cy + h/2.0)

		x0 := int(x0f)
		x1 := int(x1f)
		y0 := int(y0f)
		y1 := int(y1f)

		width := x1 - x0
		height := y1 - y0

		textBlock := fmt.Sprintf(`
            <TextBlock HPOS="%d" VPOS="%d"
                       WIDTH="%d" HEIGHT="%d"
                       ID="eSc_textblock_blck%03d"
                       TAGREFS="BT%03d">
                <Shape>
                    <Polygon POINTS="%d %d %d %d %d %d %d %d %d %d"/>
                </Shape>
            </TextBlock>`,
			x0, y0,
			width, height,
			lineIdx,
			classID,
			x0, y0, x1, y0, x1, y1, x0, y1, x0, y0,
		)

		zones = append(zones, textBlock)
		lineIdx++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return zones, nil
}

// --- helpers ---

func imageSize(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
