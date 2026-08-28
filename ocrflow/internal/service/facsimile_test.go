package service

import (
	"database/sql"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	_ "github.com/mattn/go-sqlite3"
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

func TestUpdateFromLocalDirGroupsVolumePDFsUnderOneEdition(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{
		"Paris_1615_bnf.pdf",
		"Paris_1615_vol1_bnf.pdf",
		"Paris_1615_vol1_google.PDF",
		"Paris_1615_vol2.pdf",
		"Venice_1482.pdf",
	} {
		if err := writeTestPDF(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "Paris_1615_vol3.txt"), []byte("not a PDF"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeTestPDF(filepath.Join(dir, "Unknown_1600_bnf.pdf")); err != nil {
		t.Fatal(err)
	}

	svc := newTestFacsimileService(t)
	if err := svc.UpdateFromLocalDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := svc.UpdateFromLocalDir(dir); err != nil {
		t.Fatal(err)
	}

	paris, err := svc.ListFacsimiles([]string{"Paris_1615"})
	if err != nil {
		t.Fatal(err)
	}
	if len(paris) != 4 {
		t.Fatalf("got %d Paris_1615 facsimiles, want 4", len(paris))
	}
	names := make([]string, 0, len(paris))
	for _, facsimile := range paris {
		names = append(names, filepath.Base(mustLocalPDFPath(t, facsimile)))
	}
	slices.Sort(names)
	if !slices.Equal(names, []string{"Paris_1615_bnf.pdf", "Paris_1615_vol1_bnf.pdf", "Paris_1615_vol1_google.PDF", "Paris_1615_vol2.pdf"}) {
		t.Fatalf("got PDF names %v", names)
	}

	venice, err := svc.ListFacsimiles([]string{"Venice_1482"})
	if err != nil {
		t.Fatal(err)
	}
	if len(venice) != 1 {
		t.Fatalf("got %d Venice_1482 facsimiles, want 1", len(venice))
	}

	unknown, err := svc.ListFacsimiles([]string{"Unknown_1600"})
	if err != nil {
		t.Fatal(err)
	}
	if len(unknown) != 0 {
		t.Fatalf("got %d Unknown_1600 facsimiles, want 0", len(unknown))
	}
}

func TestUpdateFromLocalDirPersistsUnambiguousShelfmarkMapping(t *testing.T) {
	dir := t.TempDir()
	if err := writeTestPDF(filepath.Join(dir, "Venice_1482.pdf")); err != nil {
		t.Fatal(err)
	}

	svc := newTestFacsimileServiceWithShelfmarks(t, []*model.EditionShelfmark{
		{ID: "shm_venice", EditionID: "Venice_1482"},
	})
	if err := svc.UpdateFromLocalDir(dir); err != nil {
		t.Fatal(err)
	}

	facsimiles, err := svc.ListFacsimiles([]string{"Venice_1482"})
	if err != nil {
		t.Fatal(err)
	}
	if len(facsimiles) != 1 {
		t.Fatalf("got %d facsimiles, want 1", len(facsimiles))
	}
	if got := facsimiles[0].ShelfmarkID; got != "shm_venice" {
		t.Fatalf("shelfmark_id = %q, want shm_venice", got)
	}
	if got := facsimiles[0].FacsimileConnectionConfirmationStatus; got != model.FacsimileConnectionStatusGuessedByMachine {
		t.Fatalf("facsimile_connection_confirmation_status = %q, want guessed_by_machine", got)
	}
}

func TestMachineGuessedShelfmarkUsesOnlyScannedShelfmark(t *testing.T) {
	got := machineGuessedShelfmark([]*model.EditionShelfmark{
		{ID: "shm_without_scan"},
		{ID: "shm_with_scan", Scan: "https://example.test/scan"},
		{ID: "shm_without_scan_2"},
	})
	if got == nil || got.ID != "shm_with_scan" {
		t.Fatalf("machineGuessedShelfmark() = %+v, want shm_with_scan", got)
	}
}

func TestMachineGuessedShelfmarkUsesOnlyGoogleBooksScan(t *testing.T) {
	got := machineGuessedShelfmark([]*model.EditionShelfmark{
		{ID: "shm_archive", Scan: "https://archive.org/details/example"},
		{ID: "shm_google", Scan: "https://www.google.com/books/edition/Foo/bar"},
		{ID: "shm_bsb", Scan: "https://www.digitale-sammlungen.de/en/view/example"},
	})
	if got == nil || got.ID != "shm_google" {
		t.Fatalf("machineGuessedShelfmark() = %+v, want shm_google", got)
	}
}

func TestMachineGuessedShelfmarkDoesNotGuessMultipleGoogleBooksScans(t *testing.T) {
	got := machineGuessedShelfmark([]*model.EditionShelfmark{
		{ID: "shm_google_1", Scan: "https://www.google.com/books/edition/Foo/one"},
		{ID: "shm_google_2", Scan: "https://books.google.com/books?id=two"},
	})
	if got != nil {
		t.Fatalf("machineGuessedShelfmark() = %+v, want no guess", got)
	}
}

func TestFacsimileEditionID(t *testing.T) {
	tests := map[string]string{
		"Paris_1615":             "Paris_1615",
		"Paris_1615_bnf":         "Paris_1615",
		"Paris_1615_vol1":        "Paris_1615",
		"Paris_1615_vol1_bnf":    "Paris_1615",
		"Paris_1615_vol1_google": "Paris_1615",
		"Paris_1615_version2":    "Paris_1615",
	}
	for fileKey, want := range tests {
		got, ok := facsimileEditionID(fileKey, []string{"Paris_1615"})
		if !ok || got != want {
			t.Errorf("facsimileEditionID(%q) = %q, %t, want %q, true", fileKey, got, ok, want)
		}
	}
	if got, ok := facsimileEditionID("Paris_1615_bnf", []string{"Paris_1615", "Paris_1615_bnf"}); !ok || got != "Paris_1615_bnf" {
		t.Errorf("facsimileEditionID() = %q, %t, want longest matching key", got, ok)
	}
	if got, ok := facsimileEditionID("Unknown_1600_bnf", []string{"Paris_1615"}); ok || got != "" {
		t.Errorf("facsimileEditionID() = %q, %t, want no match", got, ok)
	}
}

func TestUpdateFromLocalDirPreservesExistingVolumeEditionID(t *testing.T) {
	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "Paris_1615_vol1.pdf")
	if err := writeTestPDF(pdfPath); err != nil {
		t.Fatal(err)
	}
	scanURL, err := futils.LocalFilePathToURL(pdfPath)
	if err != nil {
		t.Fatal(err)
	}

	svc := newTestFacsimileService(t)
	existing := &model.Facsimile{
		EditionID: "Paris_1615_vol1",
		ScanURL:   scanURL,
	}
	existing.Name = "existing"
	if _, err := svc.CreateFacsimile(existing); err != nil {
		t.Fatal(err)
	}
	if err := svc.UpdateFromLocalDir(dir); err != nil {
		t.Fatal(err)
	}

	got, err := svc.ListFacsimiles([]string{"Paris_1615_vol1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "existing" {
		t.Fatalf("facsimiles = %#v, want the existing row to remain under Paris_1615_vol1", got)
	}
	migrated, err := svc.ListFacsimiles([]string{"Paris_1615"})
	if err != nil {
		t.Fatal(err)
	}
	if len(migrated) != 0 {
		t.Fatalf("migrated facsimiles = %#v, want no migration to Paris_1615", migrated)
	}
}

func TestUpdateFromLocalDirUpdatesPathWhenPDFDirectoryMoves(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	name := "Paris_1615_vol1.pdf"
	oldPath := filepath.Join(oldDir, name)
	newPath := filepath.Join(newDir, name)
	if err := writeTestPDF(oldPath); err != nil {
		t.Fatal(err)
	}
	if err := writeTestPDF(newPath); err != nil {
		t.Fatal(err)
	}
	oldURL, err := futils.LocalFilePathToURL(oldPath)
	if err != nil {
		t.Fatal(err)
	}

	svc := newTestFacsimileService(t)
	if _, err := svc.CreateFacsimile(&model.Facsimile{EditionID: "Paris_1615", ScanURL: oldURL}); err != nil {
		t.Fatal(err)
	}
	if err := svc.UpdateFromLocalDir(newDir); err != nil {
		t.Fatal(err)
	}

	got, err := svc.ListFacsimiles([]string{"Paris_1615"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d facsimiles after moving the PDF directory, want 1", len(got))
	}
	if path := mustLocalPDFPath(t, got[0]); path != newPath {
		t.Fatalf("PDF path = %q, want %q", path, newPath)
	}
}

func TestUpdateFacsimileRejectsDuplicateShelfmarkAssignment(t *testing.T) {
	svc := newTestFacsimileServiceWithShelfmarks(t, []*model.EditionShelfmark{
		{ID: "shm_one", EditionID: "Paris_1615"},
	})
	first, err := svc.CreateFacsimile(&model.Facsimile{EditionID: "Paris_1615", ScanURL: "https://example.test/first.pdf"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.CreateFacsimile(&model.Facsimile{EditionID: "Paris_1615", ScanURL: "https://example.test/second.pdf"})
	if err != nil {
		t.Fatal(err)
	}

	first.ShelfmarkID = "shm_one"
	first.FacsimileConnectionConfirmationStatus = model.FacsimileConnectionStatusHumanConfirmed
	if _, err := svc.UpdateFacsimile(first); err != nil {
		t.Fatal(err)
	}
	second.ShelfmarkID = "shm_one"
	second.FacsimileConnectionConfirmationStatus = model.FacsimileConnectionStatusHumanConfirmed
	if _, err := svc.UpdateFacsimile(second); err == nil {
		t.Fatal("expected duplicate shelfmark assignment to be rejected")
	}
}

func TestFacsimileMappingRecordsOnlyReportsPersistedMapping(t *testing.T) {
	records := facsimileMappingRecords(
		[]*model.Facsimile{
			{Meta: common.NewMeta("fac_one"), EditionID: "Paris_1615"},
			{Meta: common.NewMeta("fac_two"), EditionID: "Paris_1615"},
			{
				Meta:                                  common.NewMeta("fac_three"),
				EditionID:                             "Venice_1482",
				ShelfmarkID:                           "shm_venice",
				FacsimileConnectionConfirmationStatus: model.FacsimileConnectionStatusGuessedByMachine,
			},
		},
		[]*model.EditionShelfmark{
			{ID: "shm_paris", EditionID: "Paris_1615"},
			{ID: "shm_venice", EditionID: "Venice_1482"},
		},
	)
	gotParisShelfmark := records[1][9]
	gotParisStatus := records[1][10]
	gotVeniceShelfmark := records[3][9]
	gotVeniceStatus := records[3][10]

	if gotParisShelfmark != "" || gotParisStatus != "" {
		t.Fatalf("Paris row was guessed despite multiple facsimiles: shelfmark=%q status=%q", gotParisShelfmark, gotParisStatus)
	}
	if gotVeniceShelfmark != "shm_venice" || gotVeniceStatus != string(model.FacsimileConnectionStatusGuessedByMachine) {
		t.Fatalf("Venice row = shelfmark %q status %q, want persisted machine guess", gotVeniceShelfmark, gotVeniceStatus)
	}
}

func newTestFacsimileService(t *testing.T) *Facsimile {
	return newTestFacsimileServiceWithShelfmarks(t, nil)
}

func newTestFacsimileServiceWithShelfmarks(t *testing.T, shelfmarks []*model.EditionShelfmark) *Facsimile {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`
		CREATE TABLE facsimiles (
			edition_id TEXT NOT NULL,
			id TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			url TEXT NOT NULL,
			main_text_pages TEXT NOT NULL DEFAULT '',
			shelfmark_id TEXT NOT NULL DEFAULT '',
			file_size_bytes INTEGER NOT NULL DEFAULT 0,
			imported_at TIMESTAMP,
			facsimile_connection_confirmation_status TEXT NOT NULL DEFAULT '',
			PRIMARY KEY (edition_id, id)
		)
	`); err != nil {
		t.Fatal(err)
	}
	return NewFacsimileService(store.NewFacsimileSql(db, ""), func() ([]string, error) {
		return []string{"Paris_1615", "Venice_1482"}, nil
	}, func() ([]*model.EditionShelfmark, error) {
		return shelfmarks, nil
	}, "", "", "", "", "", "", "", "")
}

func mustLocalPDFPath(t *testing.T, fac *model.Facsimile) string {
	t.Helper()
	path, err := facsimileLocalPDFPath(fac)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTestPDF(path string) error {
	return os.WriteFile(path, []byte("%PDF-1.4\n"), 0o600)
}
