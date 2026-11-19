package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
)

func TestFacsimileSQLSetAvailability(t *testing.T) {
	metadataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(metadataDir, relDiagramDirs), []byte(`["Paris_1615"]`), 0o600); err != nil {
		t.Fatal(err)
	}
	pdfPath := filepath.Join(t.TempDir(), "scan.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	scanURL, err := futils.LocalFilePathToURL(pdfPath)
	if err != nil {
		t.Fatal(err)
	}

	local := &model.Facsimile{EditionID: "Paris_1615", ScanURL: scanURL}
	remote := &model.Facsimile{ScanURL: "https://example.test/scan.pdf"}
	(&FacsimileSQL{itemsMetadataDir: metadataDir}).setAvailability(remote, local)

	if remote.DownloadAvailable {
		t.Fatal("remote facsimile marked downloadable")
	}
	if !local.DownloadAvailable {
		t.Fatal("local facsimile was not marked downloadable")
	}
	if !local.DiagramCropsAvailable {
		t.Fatal("facsimile diagram crops were not marked available")
	}
}
