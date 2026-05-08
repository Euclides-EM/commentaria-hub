package store

import (
	"testing"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/csv"
)

func TestLoadEditionByKeyKeepsManuscriptElementsBooksSeparateFromBooks(t *testing.T) {
	dir := t.TempDir()

	err := csv.SaveCSV(dir+"/"+relItemsManuscript, [][]string{
		{"key", "class", "subclass", "city", "languages", "compositors", "long_title", "short_title", "short_title_source", "year_from", "year_to", "year_is_approximate", "notes", "has_diagrams"},
		{"ms_1", "Latin manuscripts", "Subclass A", "", "", "", "", "Test manuscript", "source", "1200", "1250", "True", "", ""},
	})
	if err != nil {
		t.Fatalf("save manuscript items csv: %v", err)
	}

	err = csv.SaveCSV(dir+"/"+relMDManuscript, [][]string{
		{"key", "elements_books"},
		{"ms_1", "1-3, 5"},
	})
	if err != nil {
		t.Fatalf("save manuscript metadata csv: %v", err)
	}

	store := NewEditionCSV(dir, nil)
	ed, err := store.loadEditionByKey("ms_1")
	if err != nil {
		t.Fatalf("load edition: %v", err)
	}
	if ed == nil {
		t.Fatal("expected edition")
	}

	if ed.Books != nil {
		t.Fatalf("expected books to stay nil for manuscripts, got %v", ed.Books)
	}

	if ed.ManuscriptElementsBooks == nil || *ed.ManuscriptElementsBooks != "1-3, 5" {
		t.Fatalf("expected manuscriptElementsBooks to be loaded from metadata, got %v", ed.ManuscriptElementsBooks)
	}
}

func TestLoadEditionByKeyKeepsFreeTextManuscriptElementsBooks(t *testing.T) {
	dir := t.TempDir()

	err := csv.SaveCSV(dir+"/"+relItemsManuscript, [][]string{
		{"key", "class", "subclass", "city", "languages", "compositors", "long_title", "short_title", "short_title_source", "year_from", "year_to", "year_is_approximate", "notes", "has_diagrams"},
		{"ms_60", "Latin Boethius manuscripts", "Mc", "", "", "", "", "", "", "1500", "1600", "True", "", ""},
	})
	if err != nil {
		t.Fatalf("save manuscript items csv: %v", err)
	}

	err = csv.SaveCSV(dir+"/"+relMDManuscript, [][]string{
		{"key", "elements_books"},
		{"ms_60", "copy of Naples V A 13"},
	})
	if err != nil {
		t.Fatalf("save manuscript metadata csv: %v", err)
	}

	store := NewEditionCSV(dir, nil)
	ed, err := store.loadEditionByKey("ms_60")
	if err != nil {
		t.Fatalf("load edition: %v", err)
	}
	if ed == nil {
		t.Fatal("expected edition")
	}

	if ed.ManuscriptElementsBooks == nil || *ed.ManuscriptElementsBooks != "copy of Naples V A 13" {
		t.Fatalf("expected manuscriptElementsBooks to keep free text, got %v", ed.ManuscriptElementsBooks)
	}

	if ed.Books != nil {
		t.Fatalf("expected books to stay nil for free-text manuscript elements_books, got %v", ed.Books)
	}
}

func TestLoadEditionByKeyLoadsManuscriptTitleTranslation(t *testing.T) {
	dir := t.TempDir()

	err := csv.SaveCSV(dir+"/"+relItemsManuscript, [][]string{
		{"key", "class", "subclass", "city", "languages", "compositors", "long_title", "short_title", "short_title_source", "year_from", "year_to", "year_is_approximate", "notes", "has_diagrams"},
		{"ms_1", "Latin manuscripts", "Subclass A", "", "", "", "Titulus longus", "Test manuscript", "source", "1200", "1250", "True", "", ""},
	})
	if err != nil {
		t.Fatalf("save manuscript items csv: %v", err)
	}

	err = csv.SaveCSV(dir+"/"+relTranslations, [][]string{
		{"key", "field", "en", "source"},
		{"ms_1", "title", "Long translated title", ""},
	})
	if err != nil {
		t.Fatalf("save translations csv: %v", err)
	}

	store := NewEditionCSV(dir, nil)
	ed, err := store.loadEditionByKey("ms_1")
	if err != nil {
		t.Fatalf("load edition: %v", err)
	}
	if ed == nil {
		t.Fatal("expected edition")
	}
	if ed.TitleEN == nil || *ed.TitleEN != "Long translated title" {
		t.Fatalf("expected manuscript title_EN to be loaded, got %v", ed.TitleEN)
	}
}

func TestUpsertTranslationsPersistsManuscriptTitleTranslation(t *testing.T) {
	dir := t.TempDir()

	err := csv.SaveCSV(dir+"/"+relTranslations, [][]string{
		{"key", "field", "en", "source"},
	})
	if err != nil {
		t.Fatalf("save translations csv: %v", err)
	}

	titleEN := "Long translated title"
	store := NewEditionCSV(dir, nil)
	err = store.upsertTranslations(&model.Edition{
		Key:          "ms_1",
		IsManuscript: true,
		TitleEN:      &titleEN,
	})
	if err != nil {
		t.Fatalf("upsert translations: %v", err)
	}

	_, rows, err := csv.LoadCSVRecords(dir + "/" + relTranslations)
	if err != nil {
		t.Fatalf("load translations csv: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 translation row, got %d", len(rows))
	}
	if rows[0]["key"] != "ms_1" || rows[0]["field"] != "title" || rows[0]["en"] != "Long translated title" {
		t.Fatalf("unexpected translation row: %#v", rows[0])
	}
}

func TestLoadEditionByKeyLoadsShelfmarkRepository(t *testing.T) {
	dir := t.TempDir()

	err := csv.SaveCSV(dir+"/"+relItemsPrint, [][]string{
		{"key", "short_title", "short_title_source", "year", "city", "language", "author_or_editor", "publisher", "format", "volumes", "ustc_id", "notes", "has_diagrams"},
		{"print_1", "Short title", "", "1600", "", "", "", "", "", "", "", "", ""},
	})
	if err != nil {
		t.Fatalf("save print items csv: %v", err)
	}

	err = csv.SaveCSV(dir+"/"+relShelfmarks, [][]string{
		{"key", "volume", "repository", "scan", "title_page_img", "frontispiece_img", "annotations", "shelf_mark", "copyright"},
		{"print_1", "", "Bodleian Library", "", "", "", "", "MS 1", ""},
	})
	if err != nil {
		t.Fatalf("save shelfmarks csv: %v", err)
	}

	store := NewEditionCSV(dir, nil)
	ed, err := store.loadEditionByKey("print_1")
	if err != nil {
		t.Fatalf("load edition: %v", err)
	}
	if ed == nil || len(ed.Shelfmarks) != 1 {
		t.Fatalf("expected one shelfmark, got %#v", ed)
	}
	if ed.Shelfmarks[0].Repository != "Bodleian Library" {
		t.Fatalf("expected shelfmark repository to be loaded, got %q", ed.Shelfmarks[0].Repository)
	}
}

func TestUpsertShelfmarksPersistsRepository(t *testing.T) {
	dir := t.TempDir()

	err := csv.SaveCSV(dir+"/"+relShelfmarks, [][]string{
		{"key", "volume", "repository", "scan", "title_page_img", "frontispiece_img", "annotations", "shelf_mark", "copyright"},
	})
	if err != nil {
		t.Fatalf("save shelfmarks csv: %v", err)
	}

	store := NewEditionCSV(dir, nil)
	err = store.upsertShelfmarks(&model.Edition{
		Key: "print_1",
		Shelfmarks: []model.EditionShelfmark{
			{
				Repository: "Bodleian Library",
				Shelfmark:  "MS 1",
			},
		},
	})
	if err != nil {
		t.Fatalf("upsert shelfmarks: %v", err)
	}

	_, rows, err := csv.LoadCSVRecords(dir + "/" + relShelfmarks)
	if err != nil {
		t.Fatalf("load shelfmarks csv: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 shelfmark row, got %d", len(rows))
	}
	if rows[0]["repository"] != "Bodleian Library" {
		t.Fatalf("expected persisted repository, got %#v", rows[0])
	}
}
