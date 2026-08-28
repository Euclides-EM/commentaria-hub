package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	storecsv "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/csv"
)

func TestEditionCSVUpsertShelfmarksPersistsAvailabilityMetadata(t *testing.T) {
	metadataDir := t.TempDir()
	path := filepath.Join(metadataDir, relShelfmarks)
	if err := os.WriteFile(path, []byte("key,volume,scan,title_page_img,frontispiece_img,annotations,shelf_mark,copyright,transcription_available,structural_metadata_available,note\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	volume := 1
	edition := &model.Edition{
		Key: "edition-1",
		Shelfmarks: []model.EditionShelfmark{{
			Volume:                      &volume,
			Scan:                        "https://example.test/facsimile",
			Copyright:                   "CC BY 4.0 (NOSCEMUS)",
			TranscriptionAvailable:      model.EditionShelfmarkTranscriptionExternal,
			StructuralMetadataAvailable: model.EditionShelfmarkStructuralMetadataAvailabilityInternal,
			Note:                        "External transcription",
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
	if got := rows[0]["structural_metadata_available"]; got != "internal" {
		t.Fatalf("structural_metadata_available = %q, want internal", got)
	}
	if got := rows[0]["note"]; got != "External transcription" {
		t.Fatalf("note = %q, want External transcription", got)
	}
	if got := rows[0]["copyright"]; got != "CC BY 4.0 (NOSCEMUS)" {
		t.Fatalf("copyright = %q", got)
	}
}

func TestEditionCSVUpsertShelfmarkKeepsEditionCacheWarm(t *testing.T) {
	metadataDir := t.TempDir()
	files := map[string]string{
		relItemsManuscript: "key,short_title,short_title_source,year_from,year_to,notes,has_diagrams\n",
		relItemsPrint:      "key,city,short_title,short_title_source,year,language,author_or_editor,publisher,format,volumes,ustc_id,notes,has_diagrams\nedition-1,Paris,Edition 1,,1615,French,,,,1,,,\n",
		relShelfmarks:      "id,key,volume,scan,title_page_img,frontispiece_img,annotations,shelf_mark,copyright,transcription_available,structural_metadata_available,note\nshm_1,edition-1,1,,,,,Old shelfmark,,,,\n",
	}
	for rel, contents := range files {
		if err := os.WriteFile(filepath.Join(metadataDir, rel), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	store := NewEditionCSV(metadataDir, nil)
	if err := store.WarmCache(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ListEditions(nil, nil, 0, 10); err != nil {
		t.Fatal(err)
	}

	if _, err := store.UpsertShelfmark("edition-1", &model.EditionShelfmark{
		ID:        "shm_1",
		Shelfmark: "Updated shelfmark",
	}); err != nil {
		t.Fatal(err)
	}

	ed, err := store.GetEditionByID("edition-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ed.Shelfmarks) != 1 || ed.Shelfmarks[0].Shelfmark != "Updated shelfmark" {
		t.Fatalf("cached shelfmarks = %+v, want updated shelfmark", ed.Shelfmarks)
	}
	if _, _, err := store.ListEditions(nil, nil, 0, 10); err != nil {
		t.Fatalf("ListEditions after shelfmark update failed: %v", err)
	}
}

func TestEditionShelfmarkTranscriptionAvailabilityDefaultsToNone(t *testing.T) {
	if got := editionShelfmarkTranscriptionAvailability(""); got != model.EditionShelfmarkTranscriptionNone {
		t.Fatalf("availability = %q, want none", got)
	}
}

func TestEditionShelfmarkStructuralMetadataAvailabilityDefaultsToNone(t *testing.T) {
	if got := editionShelfmarkStructuralMetadataAvailability(""); got != model.EditionShelfmarkStructuralMetadataAvailabilityNone {
		t.Fatalf("availability = %q, want none", got)
	}
}
