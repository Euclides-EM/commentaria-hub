package formatcov

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
)

func TestPageXML2ALTO(t *testing.T) {
	input := []byte(`<?xml version="1.0"?>
<PcGts xmlns="http://schema.primaresearch.org/PAGE/gts/pagecontent/2013-07-15">
  <Page imageFilename="page.jpg" imageWidth="1000" imageHeight="1500">
    <TextRegion id="r1">
      <Coords points="100,200 500,200 500,400 100,400"/>
      <TextLine id="r1l1">
        <Coords points="120,220 480,220 480,260 120,260"/>
        <Baseline points="120,255 480,255"/>
        <TextEquiv><Unicode>Hello &amp; κόσμος</Unicode></TextEquiv>
      </TextLine>
    </TextRegion>
  </Page>
</PcGts>`)

	doc, err := PageXML2ALTO(input, "page-0001")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Layout.Page) != 1 {
		t.Fatalf("got %d ALTO pages, want 1", len(doc.Layout.Page))
	}
	page := doc.Layout.Page[0]
	if page.Width != 1000 || page.Height != 1500 || page.ID != "page-0001" {
		t.Fatalf("unexpected page: %+v", page)
	}
	if doc.Description.SourceImageInformation.FileName != "page.jpg" {
		t.Fatalf("got image filename %q", doc.Description.SourceImageInformation.FileName)
	}
	block := page.PrintSpace.TextBlocks[0]
	if block.ID != "r1" || block.HPOS != 100 || block.VPOS != 200 || block.Width != 400 || block.Height != 200 {
		t.Fatalf("unexpected block: %+v", block)
	}
	line := block.Lines[0]
	if line.ID != "r1l1" || line.Baseline != "120 255 480 255" {
		t.Fatalf("unexpected line: %+v", line)
	}
	if len(line.Strings) != 1 || line.Strings[0].Content != "Hello & κόσμος" {
		t.Fatalf("unexpected strings: %+v", line.Strings)
	}
}

func TestIsPageXMLInput(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  bool
	}{
		{name: "XML directory", files: []string{"page.xml"}, want: true},
		{name: "PNG directory", files: []string{"page.png"}, want: false},
		{name: "unrelated files", files: []string{"notes.txt"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got, err := IsPageXMLInput(dir)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("IsPageXMLInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPageXMLInputRejectsMixedDirectory(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"page.xml", "page.png"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	_, err := IsPageXMLInput(dir)
	if err == nil || !strings.Contains(err.Error(), "contains both PAGE XML and PNG") {
		t.Fatalf("got error %v, want mixed-input error", err)
	}
}

func TestPageXML2ALTOUsesWordSegmentation(t *testing.T) {
	input := []byte(`<PcGts><Page imageFilename="p.png" imageWidth="100" imageHeight="200">
<TextRegion id="r"><Coords points="0,0 90,0 90,50 0,50"/><TextLine id="l">
<Coords points="0,0 90,0 90,20 0,20"/><TextEquiv><Unicode>whole line</Unicode></TextEquiv>
<Word><Coords points="0,0 30,0 30,20 0,20"/><TextEquiv><Unicode>whole</Unicode></TextEquiv></Word>
<Word><Coords points="40,0 70,0 70,20 40,20"/><TextEquiv><Unicode>line</Unicode></TextEquiv></Word>
</TextLine></TextRegion></Page></PcGts>`)

	doc, err := PageXML2ALTO(input, "p")
	if err != nil {
		t.Fatal(err)
	}
	stringsOut := doc.Layout.Page[0].PrintSpace.TextBlocks[0].Lines[0].Strings
	if len(stringsOut) != 2 || stringsOut[0].Content != "whole" || stringsOut[1].Content != "line" {
		t.Fatalf("unexpected word strings: %+v", stringsOut)
	}
}

func TestPageXMLFilesToALTO(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	input := `<PcGts><Page imageFilename="empty.jpg" imageWidth="10" imageHeight="20"/></PcGts>`
	if err := os.WriteFile(filepath.Join(inputDir, "page-0001.xml"), []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "ignore.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := PageXMLFilesToALTO(inputDir, outputDir); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(outputDir, "page-0001.xml")
	doc, err := alto.LoadFromFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Layout.Page) != 1 || doc.Layout.Page[0].ID != "page-0001" {
		t.Fatalf("unexpected converted ALTO: %+v", doc.Layout.Page)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `<alto`) || !strings.Contains(string(data), `xmlns="http://www.loc.gov/standards/alto/ns-v4#"`) {
		t.Fatalf("output is missing ALTO root/namespace: %s", data)
	}
}

func TestPageXMLFilesToALTORefusesToOverwriteInput(t *testing.T) {
	inputDir := t.TempDir()
	inputFile := filepath.Join(inputDir, "page.xml")
	input := []byte(`<PcGts><Page imageFilename="empty.jpg" imageWidth="10" imageHeight="20"/></PcGts>`)
	if err := os.WriteFile(inputFile, input, 0o644); err != nil {
		t.Fatal(err)
	}

	err := PageXMLFilesToALTO(inputDir, inputDir)
	if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("got error %v, want overwrite refusal", err)
	}
	got, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(input) {
		t.Fatal("PAGE XML input changed despite overwrite refusal")
	}
}

func TestPageXML2ALTORejectsInvalidCoordinates(t *testing.T) {
	input := []byte(`<PcGts><Page imageWidth="10" imageHeight="20"><TextRegion id="bad"><Coords points="1,2 nope"/></TextRegion></Page></PcGts>`)
	_, err := PageXML2ALTO(input, "page")
	if err == nil || !strings.Contains(err.Error(), `invalid point "nope"`) {
		t.Fatalf("got error %v, want invalid point error", err)
	}
}
