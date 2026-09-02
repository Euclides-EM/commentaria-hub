package formatcov

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
)

func TestALTOToMarkdownRendersZonesInReadingOrder(t *testing.T) {
	doc := markdownTestALTO([]markdownTestBlock{
		{category: "MainZone", text: "body", vpos: 300, hpos: 10},
		{category: "MainZone-Head--Section", text: "Methods", vpos: 200, hpos: 200},
		{category: "MainZone-Head--Book", text: "Book one", vpos: 100, hpos: 10},
		{category: "CatchWord", text: "triangle", vpos: 400, hpos: 10},
		{category: "GraphicZone-Decoration", vpos: 500, hpos: 10},
		{category: "GraphicZone-Table", vpos: 600, hpos: 10},
		{category: "Illustration", vpos: 700, hpos: 10},
		{category: "MarginTextZone-Outer", text: "a note", vpos: 800, hpos: 10},
		{category: "NumberingZone", text: "12", vpos: 900, hpos: 10},
		{category: "UnknownZone", text: "unknown text", vpos: 1000, hpos: 10},
		// Same vertical position as the section, but farther left.
		{category: "RunningTitleZone", text: "running", vpos: 200, hpos: 10},
	})

	want := "# Book one\n" +
		"<!-- Running title: running -->\n" +
		"## Methods\n" +
		"body\n" +
		"<!-- Catchword: triangle -->\n" +
		"*[Ornament]*\n" +
		"*[Figure: Table]*\n" +
		"*[Figure]*\n" +
		"*[Margin: a note]*\n" +
		"<!-- Page: 12 -->\n" +
		"unknown text\n"
	if got := ALTOToMarkdown(doc); got != want {
		t.Fatalf("ALTOToMarkdown() =\n%q\nwant:\n%q", got, want)
	}
}

func TestALTOToMarkdownEmptyZoneRulesAndNormalization(t *testing.T) {
	doc := markdownTestALTO([]markdownTestBlock{
		{category: "DigitizationArtefactZone", vpos: 10},
		{category: "DropCapitalZone", text: "A", vpos: 20},
		{category: "DropCapitalZone-Plain", vpos: 30},
		{category: "QuireMarksZone", text: "B2", vpos: 40},
		{category: "MarginTextZone", vpos: 50},
		{category: "NumberingZone", vpos: 60},
		{category: "UnknownZone", vpos: 70},
	})

	want := "<!-- Digitization artefact -->\n" +
		"<!-- Drop capital: A -->\n" +
		"<!-- Drop capital (plain) -->\n" +
		"<!-- Quire marks: B2 -->\n"
	if got := ALTOToMarkdown(doc); got != want {
		t.Fatalf("ALTOToMarkdown() =\n%q\nwant:\n%q", got, want)
	}
}

func TestALTOFilesToMarkdown(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	inputPath := filepath.Join(inputDir, "page-0001.xml")
	if err := alto.SaveToFile(markdownTestALTO([]markdownTestBlock{{category: "MainZone", text: "hello"}}), inputPath); err != nil {
		t.Fatal(err)
	}

	if err := ALTOFilesToMarkdown(inputDir, outputDir); err != nil {
		t.Fatalf("ALTOFilesToMarkdown() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(outputDir, "page-0001.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello\n" {
		t.Fatalf("output = %q, want %q", got, "hello\\n")
	}
}

type markdownTestBlock struct {
	category string
	text     string
	vpos     float64
	hpos     float64
}

func markdownTestALTO(blocks []markdownTestBlock) *alto.Alto {
	doc := &alto.Alto{Layout: alto.Layout{Page: []alto.Page{{}}}}
	for i, input := range blocks {
		tagID := "tag-" + string(rune('a'+i))
		doc.Tags.OtherTags = append(doc.Tags.OtherTags, alto.OtherTag{ID: tagID, Label: input.category})
		block := alto.TextBlock{TagRefs: tagID, VPOS: input.vpos, HPOS: input.hpos}
		if input.text != "" {
			block.Lines = []alto.TextLine{{Strings: []alto.String{{Content: input.text}}}}
		}
		doc.Layout.Page[0].PrintSpace.TextBlocks = append(doc.Layout.Page[0].PrintSpace.TextBlocks, block)
	}
	return doc
}
