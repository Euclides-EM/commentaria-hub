package tei

import (
	"log"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
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
	txtPath := path.Join(testDataDir, td, transcriptionTXTFilename)
	txtBytes, err := os.ReadFile(txtPath)
	if err != nil {
		t.Fatalf("failed to read TXT file for test %s: %v", td, err)
	}
	txt := string(txtBytes)
	linesInput := LinesInput{
		LinesByKeys: map[string]Lines{
			"page1": {
				TranscriptionLines: strings.Split(txt, "\n"),
			},
		},
	}

	tei, err := BuildTEIFromLines(linesInput, EntitiesByTest[td], ImgURLsByTest[td])
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

	expectedTEIPath := path.Join(testDataDir, td, expectedTEIXMLFilename)
	expectedTEIBytes, err := os.ReadFile(expectedTEIPath)
	if err != nil {
		t.Fatalf("failed to read expected TEI XML file for test %s: %v", td, err)
	}

	expectedTEI, err := ParseTEIFromXML(expectedTEIBytes)
	if err != nil {
		t.Fatalf("failed to parse expected TEI XML for test %s: %v", td, err)
	}

	if !reflect.DeepEqual(tei, expectedTEI) {
		t.Fatalf("TEI output does not match expected for test %s", td)
	}
}
