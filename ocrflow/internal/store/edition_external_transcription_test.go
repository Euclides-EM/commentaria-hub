package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
)

func TestEditionCSVExternalTranscriptionsRoundTrip(t *testing.T) {
	metadataDir := t.TempDir()
	writeEditionCSVFixture(t, metadataDir, relItemsManuscript, "key,short_title\nms_1,Test manuscript\n")
	writeEditionCSVFixture(t, metadataDir, relExternalTranscriptions, "key,url,note\n")

	store := NewEditionCSV(metadataDir, func(string) {})
	edition := &model.Edition{
		Key: "ms_1",
		ExternalTranscriptions: []model.ExternalTranscription{
			{URL: "https://example.org/transcription/1", Note: "Diplomatic transcription"},
			{URL: "https://example.org/transcription/2", Note: "Normalized transcription"},
		},
	}

	if err := store.upsertExternalTranscriptions(edition); err != nil {
		t.Fatalf("upsert external transcriptions: %v", err)
	}
	loaded, err := store.loadEditionByKey(edition.Key)
	if err != nil {
		t.Fatalf("load edition: %v", err)
	}
	if got, want := len(loaded.ExternalTranscriptions), 2; got != want {
		t.Fatalf("external transcription count = %d, want %d", got, want)
	}
	if got, want := loaded.ExternalTranscriptions[0].URL, edition.ExternalTranscriptions[0].URL; got != want {
		t.Fatalf("external transcription URL = %q, want %q", got, want)
	}
	if got, want := loaded.ExternalTranscriptions[1].Note, edition.ExternalTranscriptions[1].Note; got != want {
		t.Fatalf("external transcription note = %q, want %q", got, want)
	}

	edition.ExternalTranscriptions = nil
	if err := store.upsertExternalTranscriptions(edition); err != nil {
		t.Fatalf("delete external transcriptions: %v", err)
	}
	loaded, err = store.loadEditionByKey(edition.Key)
	if err != nil {
		t.Fatalf("reload edition: %v", err)
	}
	if len(loaded.ExternalTranscriptions) != 0 {
		t.Fatalf("external transcriptions were not deleted: %#v", loaded.ExternalTranscriptions)
	}
}

func TestBuildEditionFromPreloadedIncludesExternalTranscriptions(t *testing.T) {
	store := NewEditionCSV(t.TempDir(), func(string) {})
	loaded := store.buildEditionFromPreloaded("print_1", &preloadedEditionRows{
		printRows: []map[string]string{{"key": "print_1"}},
		externalTranscriptions: []map[string]string{{
			"key":  "print_1",
			"url":  "https://example.org/transcription/1",
			"note": "Machine-readable text",
		}, {
			"key":  "print_1",
			"url":  "https://example.org/transcription/2",
			"note": "Second text",
		}},
	})

	if loaded == nil || len(loaded.ExternalTranscriptions) != 2 {
		t.Fatalf("external transcriptions were not included in bulk-loaded edition: %#v", loaded)
	}
	if got, want := loaded.ExternalTranscriptions[0].Note, "Machine-readable text"; got != want {
		t.Fatalf("external transcription note = %q, want %q", got, want)
	}
}

func writeEditionCSVFixture(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
