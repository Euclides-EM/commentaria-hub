package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
)

func TestDownloadableFacsimilePDFPathSkipsUnavailableFacsimile(t *testing.T) {
	pdfPath := filepath.Join(t.TempDir(), "main.pdf")
	if err := writeTestPDF(pdfPath); err != nil {
		t.Fatal(err)
	}
	scanURL, err := futils.LocalFilePathToURL(pdfPath)
	if err != nil {
		t.Fatal(err)
	}

	got, err := downloadableFacsimilePDFPath([]*model.Facsimile{
		{ScanURL: "https://example.test/scan.pdf"},
		{ScanURL: scanURL},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != pdfPath {
		t.Fatalf("downloadableFacsimilePDFPath() = %q, want %q", got, pdfPath)
	}
}

func writeTestPDF(path string) error {
	return os.WriteFile(path, []byte("%PDF-1.4\n"), 0o600)
}
