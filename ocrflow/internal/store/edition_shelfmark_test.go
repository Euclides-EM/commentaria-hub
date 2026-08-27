package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	storecsv "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/csv"
)

func TestEditionCSVUpsertShelfmarksPersistsTranscriptionMetadata(t *testing.T) {
	metadataDir := t.TempDir()
	path := filepath.Join(metadataDir, relShelfmarks)
	if err := os.WriteFile(path, []byte("key,volume,scan,title_page_img,frontispiece_img,annotations,shelf_mark,copyright,transcription_available,note\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	volume := 1
	edition := &model.Edition{
		Key: "edition-1",
		Shelfmarks: []model.EditionShelfmark{{
			Volume:                 &volume,
			Scan:                   "https://example.test/facsimile",
			Copyright:              "CC BY 4.0 (NOSCEMUS)",
			TranscriptionAvailable: model.EditionShelfmarkTranscriptionExternal,
			Note:                   "External transcription",
		}},
	}

	if err := NewEditionCSV(metadataDir, nil).upsertShelfmarks(edition); err != nil {
		t.Fatal(err)
	}
	_, rows, err := storecsv.LoadCSVRecords(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d shelfmark rows, want 1", len(rows))
	}
	if got := rows[0]["transcription_available"]; got != "external" {
		t.Fatalf("transcription_available = %q, want external", got)
	}
	if got := rows[0]["note"]; got != "External transcription" {
		t.Fatalf("note = %q, want External transcription", got)
	}
	if got := rows[0]["copyright"]; got != "CC BY 4.0 (NOSCEMUS)" {
		t.Fatalf("copyright = %q", got)
	}
}

func TestEditionShelfmarkTranscriptionAvailabilityDefaultsToNone(t *testing.T) {
	if got := editionShelfmarkTranscriptionAvailability(""); got != model.EditionShelfmarkTranscriptionNone {
		t.Fatalf("availability = %q, want none", got)
	}
}
