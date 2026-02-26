package tei

import (
	"log"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
)

const (
	testDataDir               = "testdata"
	transcriptionTXTFilename  = "original.txt"
	transcriptionALTOFilename = "input.xml"
	expectedTEIXMLFilename    = "expected.xml"

	defaultPageKey = "page1"
)

func TestBuildTEI(t *testing.T) {
	des, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	var testDirs []string
	for _, de := range des {
		if de.IsDir() {
			testDirs = append(testDirs, de.Name())
		}
	}

	for _, td := range testDirs {
		t.Run(td, func(t *testing.T) {
			_, err := os.Stat(path.Join(testDataDir, td, transcriptionALTOFilename))
			if err != nil && !os.IsNotExist(err) {
				t.Fatalf("ALTO file not found for test %s: %v", td, err)
			}
			if err == nil {
				testBuildTEIFromALTO(t, td)
				return
			}
			_, err = os.Stat(path.Join(testDataDir, td, transcriptionTXTFilename))
			if err != nil && os.IsNotExist(err) {
				t.Fatalf("TXT file not found for test %s: %v", td, err)
			}
			if err == nil {
				testBuildTEIFromTXT(t, td)
				return
			}
			t.Fatalf("no valid transcription file found for test %s", td)
		})
	}
}

func testBuildTEIFromTXT(t *testing.T, td string) {
	tdPath := path.Join(testDataDir, td)
	txtPath := path.Join(tdPath, transcriptionTXTFilename)
	txtBytes, err := os.ReadFile(txtPath)
	if err != nil {
		t.Fatalf("failed to read TXT file for test %s: %v", td, err)
	}
	txt := strings.TrimRight(string(txtBytes), "\n")
	transcriptionLines := strings.Split(txt, "\n")
	lines := Lines{
		TranscriptionLines: transcriptionLines,
		Translations:       map[string][]string{},
	}
	// Load optional translation files (en.txt, fr.txt, etc.)
	des, err := os.ReadDir(tdPath)
	if err != nil {
		t.Fatalf("failed to read testdata dir %s: %v", td, err)
	}
	for _, de := range des {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".txt") {
			continue
		}
		if de.Name() == transcriptionTXTFilename {
			continue
		}
		lang := strings.TrimSuffix(de.Name(), ".txt")
		trPath := path.Join(tdPath, de.Name())
		trBytes, err := os.ReadFile(trPath)
		if err != nil {
			t.Fatalf("failed to read translation file %s: %v", trPath, err)
		}
		trContent := strings.TrimRight(string(trBytes), "\n")
		lines.Translations[lang] = strings.Split(trContent, "\n")
	}
	imgURL := ""
	if m, ok := ImgURLsByTest[td]; ok {
		imgURL = m[defaultPageKey]
	}
	tei, err := BuildTEIFromLines(defaultPageKey, lines, EntitiesByTest[td], imgURL)
	if err != nil {
		t.Fatalf("failed to build TEI from lines for test %s: %v", td, err)
	}

	actualTEIBytes, err := tei.ToXML()
	if err != nil {
		t.Fatalf("failed to serialize actual TEI to XML for test %s: %v", td, err)
	}

	expectedTEIPath := path.Join(testDataDir, td, expectedTEIXMLFilename)
	expectedTEIBytes, err := os.ReadFile(expectedTEIPath)
	if err != nil {
		t.Fatalf("failed to read expected TEI XML file for test %s: %v", td, err)
	}

	actualStr := strings.TrimSpace(string(actualTEIBytes))
	expectedStr := strings.TrimSpace(string(expectedTEIBytes))
	if actualStr != expectedStr {
		log.Printf("expected TEI XML for test %s:\n"+
			"-----------------------\n"+
			"%v\n"+
			"-----------------------\n"+
			"%v", td, expectedStr, actualStr)
		t.Fatalf("TEI XML output does not match expected for test %s", td)
	}
}

func testBuildTEIFromALTO(t *testing.T, td string) {
	altoPath := path.Join(testDataDir, td, transcriptionALTOFilename)
	a, err := alto.LoadFromFile(altoPath)
	if err != nil {
		t.Fatalf("failed to load ALTO file for test %s: %v", td, err)
	}
	ImgURL := ""
	if ImgURLMap, ok := ImgURLsByTest[td]; ok {
		ImgURL = ImgURLMap[defaultPageKey]
	}
	tei, err := BuildTEIFromALTO(a, EntitiesByTest[td], ImgURL)
	if err != nil {
		t.Fatalf("failed to build TEI from ALTO for test %s: %v", td, err)
	}

	actualTEIBytes, err := tei.ToXML()
	if err != nil {
		t.Fatalf("failed to serialize TEI to XML for test %s: %v", td, err)
	}

	expectedTEIPath := path.Join(testDataDir, td, expectedTEIXMLFilename)
	expectedTEIBytes, err := os.ReadFile(expectedTEIPath)
	if err != nil {
		t.Fatalf("failed to read expected TEI XML file for test %s: %v", td, err)
	}

	actualStr := strings.TrimSpace(string(actualTEIBytes))
	expectedStr := strings.TrimSpace(string(expectedTEIBytes))
	if actualStr != expectedStr {
		log.Printf("expected TEI XML for test %s:\n"+
			"-----------------------\n"+
			"%v\n"+
			"-----------------------\n"+
			"%v", td, expectedStr, actualStr)
		t.Fatalf("TEI XML output does not match expected for test %s", td)
	}
}

// TestParseTEIFromXML_valid verifies that parsing valid TEI XML and re-serializing produces equivalent TEI.
func TestParseTEIFromXML_valid(t *testing.T) {
	expectedPath := path.Join(testDataDir, "lines_single_entity_translation", expectedTEIXMLFilename)
	xmlBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read expected TEI: %v", err)
	}
	tei, err := ParseTEIFromXML(xmlBytes)
	if err != nil {
		t.Fatalf("ParseTEIFromXML: %v", err)
	}
	if tei == nil {
		t.Fatal("ParseTEIFromXML returned nil TEI")
	}
	out, err := tei.ToXML()
	if err != nil {
		t.Fatalf("ToXML: %v", err)
	}
	tei2, err := ParseTEIFromXML(out)
	if err != nil {
		t.Fatalf("ParseTEIFromXML(roundtrip): %v", err)
	}
	if !reflect.DeepEqual(tei, tei2) {
		t.Error("TEI roundtrip: parsed → ToXML → Parse produced different struct")
	}
}

// TestParseTEIFromXML_invalid verifies that invalid XML returns an error.
func TestParseTEIFromXML_invalid(t *testing.T) {
	_, err := ParseTEIFromXML([]byte("not xml <<<"))
	if err == nil {
		t.Error("ParseTEIFromXML with invalid XML should return error")
	}
}

// TestTEI_ToXML_nil verifies that calling ToXML on a nil TEI returns an error.
func TestTEI_ToXML_nil(t *testing.T) {
	var doc *model.TEI
	_, err := doc.ToXML()
	if err == nil {
		t.Error("ToXML on nil TEI should return error")
	}
}

// TestBuildTEIFromLines_emptyInput builds TEI from empty lines for one page and checks valid output without panic.
func TestBuildTEIFromLines_emptyInput(t *testing.T) {
	tei, err := BuildTEIFromLines("page1", Lines{TranscriptionLines: nil, Translations: nil}, nil, "")
	if err != nil {
		t.Fatalf("BuildTEIFromLines(empty): %v", err)
	}
	if tei == nil {
		t.Fatal("BuildTEIFromLines returned nil")
	}
	// One page with no lines still produces one surface and one (empty) div
	if len(tei.Facsimile.Surfaces) != 1 {
		t.Errorf("expected 1 surface, got %d", len(tei.Facsimile.Surfaces))
	}
	if len(tei.Text.Body.Divs) != 1 {
		t.Errorf("expected 1 div, got %d", len(tei.Text.Body.Divs))
	}
	out, err := tei.ToXML()
	if err != nil {
		t.Fatalf("ToXML: %v", err)
	}
	if len(out) == 0 {
		t.Error("ToXML produced empty output")
	}
}

// TestBuildTEIFromLines_noEntities verifies that TEI built without entities has no encodingDesc or standOff.
func TestBuildTEIFromLines_noEntities(t *testing.T) {
	lines := Lines{
		TranscriptionLines: []string{"First line.", "Second line."},
	}
	tei, err := BuildTEIFromLines("page1", lines, nil, "")
	if err != nil {
		t.Fatalf("BuildTEIFromLines: %v", err)
	}
	if tei.Header.EncodingDesc != nil {
		t.Error("expected no encodingDesc when entities is nil")
	}
	if tei.Header.StandOff != nil {
		t.Error("expected no standOff when entities is nil")
	}
	// One surface, one transcription div
	if n := len(tei.Facsimile.Surfaces); n != 1 {
		t.Errorf("expected 1 surface, got %d", n)
	}
	if n := len(tei.Text.Body.Divs); n != 1 {
		t.Errorf("expected 1 div (transcription), got %d", n)
	}
	// Body text should contain the two lines (as CharData + LB nodes)
	abs := tei.Text.Body.Divs[0].Abs
	if len(abs) != 1 {
		t.Fatalf("expected 1 ab, got %d", len(abs))
	}
	nodes := abs[0].Nodes
	var textParts []string
	for _, nd := range nodes {
		if nd.CharData != "" {
			textParts = append(textParts, nd.CharData)
		}
	}
	combined := strings.Join(textParts, "")
	if !strings.Contains(combined, "First line.") || !strings.Contains(combined, "Second line.") {
		t.Errorf("body should contain both lines, got: %q", combined)
	}
}

// TestBuildTEIFromLines_multiplePages verifies that one page per call produces one surface and one div each.
func TestBuildTEIFromLines_multiplePages(t *testing.T) {
	tei1, err := BuildTEIFromLines("page1", Lines{TranscriptionLines: []string{"Page one."}}, nil, "")
	if err != nil {
		t.Fatalf("BuildTEIFromLines(page1): %v", err)
	}
	tei2, err := BuildTEIFromLines("page2", Lines{TranscriptionLines: []string{"Page two."}}, nil, "")
	if err != nil {
		t.Fatalf("BuildTEIFromLines(page2): %v", err)
	}
	for _, tc := range []struct {
		name string
		tei  *model.TEI
		page string
	}{
		{"page1", tei1, "page1"},
		{"page2", tei2, "page2"},
	} {
		if n := len(tc.tei.Facsimile.Surfaces); n != 1 {
			t.Errorf("%s: expected 1 surface, got %d", tc.name, n)
		}
		if n := len(tc.tei.Text.Body.Divs); n != 1 {
			t.Errorf("%s: expected 1 div, got %d", tc.name, n)
		}
		if tc.tei.Facsimile.Surfaces[0].N != tc.page {
			t.Errorf("%s: expected surface n=%q, got %q", tc.name, tc.page, tc.tei.Facsimile.Surfaces[0].N)
		}
	}
}

// TestBuildTEIFromLines_emptyImageUrl verifies that empty imageUrl does not panic and facs is empty.
func TestBuildTEIFromLines_emptyImageUrl(t *testing.T) {
	tei, err := BuildTEIFromLines("page1", Lines{TranscriptionLines: []string{"One line."}}, nil, "")
	if err != nil {
		t.Fatalf("BuildTEIFromLines: %v", err)
	}
	if tei.Facsimile.Surfaces[0].Facs != "" {
		t.Errorf("expected empty facs when imageUrl is empty, got %q", tei.Facsimile.Surfaces[0].Facs)
	}
}

// TestBuildTEIFromLines_utf8EntitySpan verifies entity spanning multi-byte UTF-8 runes is sliced safely.
func TestBuildTEIFromLines_utf8EntitySpan(t *testing.T) {
	// "café" has 'é' as 2 bytes; entity "fé" (bytes 3–5) must not break UTF-8
	line := "café"
	entities := []EntityItem{
		{
			Ref:   "ent_1",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 3}, // start of 'é'
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 5}, // end of 'é' (2 bytes)
		},
	}
	tei, err := BuildTEIFromLines("page1", Lines{TranscriptionLines: []string{line}}, entities, "")
	if err != nil {
		t.Fatalf("BuildTEIFromLines: %v", err)
	}
	// Body should contain the entity text; safeSliceByByteRange should yield "é"
	abs := tei.Text.Body.Divs[0].Abs
	var found string
	for _, nd := range abs[0].Nodes {
		if nd.CharData != "" {
			found += nd.CharData
		}
	}
	if !strings.Contains(found, "é") {
		t.Errorf("expected body to contain entity slice é, got %q", found)
	}
	if !strings.Contains(found, "caf") {
		t.Errorf("expected body to contain prefix caf, got %q", found)
	}
}

// TestBuildTEIFromALTO_nilAlto verifies that nil ALTO returns an error.
func TestBuildTEIFromALTO_nilAlto(t *testing.T) {
	_, err := BuildTEIFromALTO(nil, nil, "")
	if err == nil {
		t.Error("BuildTEIFromALTO(nil) should return error")
	}
}

// TestBuildTEIFromALTO_minimal verifies TEI from a minimal in-memory ALTO (one page, one block, one line).
func TestBuildTEIFromALTO_minimal(t *testing.T) {
	a := &alto.Alto{
		Layout: alto.Layout{
			Page: []alto.Page{{
				ID: "p1",
				PrintSpace: alto.PrintSpace{
					TextBlocks: []alto.TextBlock{{
						ID: "b1",
						Lines: []alto.TextLine{{
							Strings: []alto.AltoString{{Content: "Hello"}, {Content: "world."}},
						}},
					}},
				},
			}},
		},
	}
	tei, err := BuildTEIFromALTO(a, nil, "https://example.com/p1.png")
	if err != nil {
		t.Fatalf("BuildTEIFromALTO: %v", err)
	}
	if tei == nil {
		t.Fatal("BuildTEIFromALTO returned nil")
	}
	if n := len(tei.Facsimile.Surfaces); n != 1 {
		t.Errorf("expected 1 surface, got %d", n)
	}
	if tei.Facsimile.Surfaces[0].Facs != "https://example.com/p1.png" {
		t.Errorf("expected facs URL on surface, got %q", tei.Facsimile.Surfaces[0].Facs)
	}
	if n := len(tei.Text.Body.Divs); n != 1 {
		t.Errorf("expected 1 div, got %d", n)
	}
	segs := tei.Text.Body.Divs[0].Abs[0].Segs
	if len(segs) != 1 {
		t.Fatalf("expected 1 seg, got %d", len(segs))
	}
	// Line text is joined with space: "Hello" + " " + "world."
	var lineText string
	for _, in := range segs[0].Content {
		lineText += in.Text
	}
	if lineText != "Hello world." {
		t.Errorf("expected line text 'Hello world.', got %q", lineText)
	}
}
