package formatcov

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
)

var commentZoneNames = map[string]string{
	"CatchWord":                "Catchword",
	"DigitizationArtefactZone": "",
	"DropCapitalZone":          "",
	"DropCapitalZone-Plain":    "",
	"RunningTitleZone":         "",
	"QuireMarksZone":           "",
}

type markdownBlock struct {
	pageIndex  int
	blockIndex int
	vpos       float64
	hpos       float64
	category   string
	content    string
}

// ALTOFilesToMarkdown converts an ALTO XML file or every ALTO XML file in a
// flat directory. Output filenames use the input basename with a .md suffix.
func ALTOFilesToMarkdown(inputPath, outputDir string) error {
	paths, err := flatInputFiles(inputPath, ".xml", "ALTO XML")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create Markdown output directory %q: %w", outputDir, err)
	}

	for _, inputFile := range paths {
		doc, err := alto.LoadFromFile(inputFile)
		if err != nil {
			return fmt.Errorf("load %q: %w", inputFile, err)
		}
		markdown := ALTOToMarkdown(doc)
		base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
		outputFile := filepath.Join(outputDir, base+".md")
		fmt.Printf("converting ALTO %q -> %q\n", inputFile, outputFile)
		if err := os.WriteFile(outputFile, []byte(markdown), 0o644); err != nil {
			return fmt.Errorf("write Markdown %q: %w", outputFile, err)
		}
	}
	return nil
}

// ALTOToMarkdown renders ALTO text blocks in page order, then from top to
// bottom and left to right within each page.
func ALTOToMarkdown(doc *alto.Alto) string {
	if doc == nil {
		return ""
	}

	tagLabels := make(map[string]string, len(doc.Tags.OtherTags))
	for _, tag := range doc.Tags.OtherTags {
		if tag.ID != "" {
			tagLabels[tag.ID] = tag.Label
		}
	}

	var blocks []markdownBlock
	for pageIndex, page := range doc.Layout.Page {
		for blockIndex := range page.PrintSpace.TextBlocks {
			block := &page.PrintSpace.TextBlocks[blockIndex]
			blocks = append(blocks, markdownBlock{
				pageIndex:  pageIndex,
				blockIndex: blockIndex,
				vpos:       block.VPOS,
				hpos:       block.HPOS,
				category:   firstTagLabel(block.TagRefs, tagLabels),
				content:    alto.ExtractTextContentFromBlock(block),
			})
		}
	}

	sort.SliceStable(blocks, func(i, j int) bool {
		if blocks[i].pageIndex != blocks[j].pageIndex {
			return blocks[i].pageIndex < blocks[j].pageIndex
		}
		if blocks[i].vpos != blocks[j].vpos {
			return blocks[i].vpos < blocks[j].vpos
		}
		if blocks[i].hpos != blocks[j].hpos {
			return blocks[i].hpos < blocks[j].hpos
		}
		return blocks[i].blockIndex < blocks[j].blockIndex
	})

	lines := make([]string, 0, len(blocks))
	for _, block := range blocks {
		if rendered := renderMarkdownBlock(block.category, block.content); rendered != "" {
			lines = append(lines, rendered)
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderMarkdownBlock(category, content string) string {
	content = strings.TrimSpace(content)

	if override, ok := commentZoneNames[category]; ok {
		name := override
		if name == "" {
			name = normalizeZoneName(category)
		}
		if content == "" {
			return fmt.Sprintf("<!-- %s -->", name)
		}
		return fmt.Sprintf("<!-- %s: %s -->", name, content)
	}

	switch category {
	case "GraphicZone-Decoration":
		return "*[Ornament]*"
	case "Illustration":
		return "*[Figure]*"
	case "MainZone-Head--Book":
		return markdownHeading(1, content)
	case "MainZone-Head--Section":
		return markdownHeading(2, content)
	case "NumberingZone":
		if content == "" {
			return ""
		}
		return fmt.Sprintf("<!-- Page: %s -->", content)
	}

	if category == "MarginTextZone" || strings.HasPrefix(category, "MarginTextZone-") {
		if content == "" {
			return ""
		}
		return fmt.Sprintf("*[Margin: %s]*", content)
	}
	if category == "GraphicZone" {
		return "*[Figure]*"
	}
	if strings.HasPrefix(category, "GraphicZone-") {
		return fmt.Sprintf("*[Figure: %s]*", strings.TrimPrefix(category, "GraphicZone-"))
	}

	return content
}

func markdownHeading(level int, content string) string {
	if content == "" {
		return ""
	}
	return strings.Repeat("#", level) + " " + content
}

func firstTagLabel(tagRefs string, labels map[string]string) string {
	for _, ref := range strings.Fields(tagRefs) {
		if label, ok := labels[ref]; ok {
			return label
		}
	}
	return ""
}

func normalizeZoneName(category string) string {
	parts := strings.Split(category, "-")
	base := strings.TrimSuffix(parts[0], "Zone")
	words := splitIdentifierWords(base)
	name := strings.ToLower(strings.Join(words, " "))
	if name != "" {
		runes := []rune(name)
		runes[0] = unicode.ToUpper(runes[0])
		name = string(runes)
	}
	if len(parts) > 1 {
		qualifier := strings.ToLower(strings.Join(parts[1:], " "))
		name += " (" + qualifier + ")"
	}
	return name
}

func splitIdentifierWords(value string) []string {
	if value == "" {
		return nil
	}
	runes := []rune(value)
	start := 0
	var words []string
	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) && (unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
			words = append(words, string(runes[start:i]))
			start = i
		}
	}
	return append(words, string(runes[start:]))
}

func flatInputFiles(inputPath, extension, description string) ([]string, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s input %q: %w", description, inputPath, err)
	}
	if !info.IsDir() {
		if !strings.EqualFold(filepath.Ext(inputPath), extension) {
			return nil, fmt.Errorf("%s input %q must be a %s file or directory", description, inputPath, extension)
		}
		return []string{inputPath}, nil
	}

	entries, err := os.ReadDir(inputPath)
	if err != nil {
		return nil, fmt.Errorf("read %s directory %q: %w", description, inputPath, err)
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), extension) {
			paths = append(paths, filepath.Join(inputPath, entry.Name()))
		}
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("%s directory %q has no %s files", description, inputPath, extension)
	}
	return paths, nil
}
