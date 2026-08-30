package krakenwrapper

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestKrakenOCRReuseAltoArgsUsesExistingALTO(t *testing.T) {
	pairs := [][2]string{{"/tmp/page.png", "/tmp/page.xml"}}
	got := krakenOCRReuseAltoArgs(pairs, "/tmp/model.mlmodel")
	want := []string{
		"--alto", "--format-type", "alto", "--raise-on-error",
		"--device", krakenDeviceArg(),
		"-i", "/tmp/page.xml", "/tmp/page.xml.ocr.tmp",
		"ocr", "-m", "/tmp/model.mlmodel",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Kraken OCR args = %q, want %q", got, want)
	}
}

func TestPrepareAltoForOCR(t *testing.T) {
	path := filepath.Join(t.TempDir(), "page.xml")
	input := `<alto xmlns="http://www.loc.gov/standards/alto/ns-v4#"><Description><sourceImageInformation><fileName>old.png</fileName></sourceImageInformation></Description><Layout><Page><PrintSpace><TextBlock ID="empty"/><TextBlock ID="content" HPOS="0" VPOS="0" WIDTH="100" HEIGHT="100"><TextLine ID="line-1" HPOS="10" VPOS="20" WIDTH="30" HEIGHT="40" BASELINE="10 50 40 50"/></TextBlock></PrintSpace></Page></Layout></alto>`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := prepareAltoForOCR(path, "page.png"); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(output)
	for _, expected := range []string{
		"<fileName>page.png</fileName>",
		`<Polygon POINTS="10 20 40 20 40 60 10 60 10 20"/>`,
	} {
		if !strings.Contains(text, expected) {
			t.Errorf("prepared ALTO does not contain %q: %s", expected, text)
		}
	}
	if strings.Contains(text, `ID="empty"`) {
		t.Errorf("prepared ALTO retains invalid empty block: %s", text)
	}
}

func TestPrepareAltoForOCRRejectsLineWithoutGeometry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "page.xml")
	input := `<alto><Description><sourceImageInformation><fileName>old.png</fileName></sourceImageInformation></Description><Layout><Page><PrintSpace><TextBlock><TextLine ID="line-1" BASELINE="10 50 40 50"/></TextBlock></PrintSpace></Page></Layout></alto>`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}

	err := prepareAltoForOCR(path, "page.png")
	if err == nil || !strings.Contains(err.Error(), `text line "line-1" has no valid polygon or bounding box`) {
		t.Fatalf("prepareAltoForOCR error = %v", err)
	}
}
