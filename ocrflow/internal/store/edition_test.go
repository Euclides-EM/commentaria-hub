package store

import (
	"testing"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/csv"
)

func TestLoadEditionByKeyLoadsManuscriptElementsMetadataLikePrint(t *testing.T) {
	dir := t.TempDir()

	err := csv.SaveCSV(dir+"/"+relItemsManuscript, [][]string{
		{"key", "class", "subclass", "city", "languages", "compositors", "long_title", "short_title", "short_title_source", "year_from", "year_to", "year_is_approximate", "notes", "has_diagrams"},
		{"ms_1", "Latin manuscripts", "Subclass A", "", "", "", "", "Test manuscript", "source", "1200", "1250", "True", "", ""},
	})
	if err != nil {
		t.Fatalf("save manuscript items csv: %v", err)
	}

	err = csv.SaveCSV(dir+"/"+relMDManuscript, [][]string{
		{"key", "elements_books", "additional_content", "elements_content"},
		{"ms_1", "1-3, 5", "scholia, diagrams", "I-XV with scholia"},
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

	if len(ed.Books) != 4 || ed.Books[0] != 1 || ed.Books[1] != 2 || ed.Books[2] != 3 || ed.Books[3] != 5 {
		t.Fatalf("expected manuscript books to be parsed from metadata, got %v", ed.Books)
	}

	if len(ed.AdditionalContent) != 2 || ed.AdditionalContent[0] != "scholia" || ed.AdditionalContent[1] != "diagrams" {
		t.Fatalf("expected manuscript additional content to be parsed from metadata, got %v", ed.AdditionalContent)
	}

	if ed.ManuscriptElementsContent == nil || *ed.ManuscriptElementsContent != "I-XV with scholia" {
		t.Fatalf("expected manuscript elements content to be loaded from metadata, got %v", ed.ManuscriptElementsContent)
	}
}

func TestUpsertManuscriptPersistsElementsMetadataLikePrint(t *testing.T) {
	dir := t.TempDir()

	err := csv.SaveCSV(dir+"/"+relItemsManuscript, [][]string{
		{"key", "class", "subclass", "city", "languages", "compositors", "long_title", "short_title", "short_title_source", "year_from", "year_to", "year_is_approximate", "notes", "has_diagrams"},
	})
	if err != nil {
		t.Fatalf("save manuscript items csv: %v", err)
	}

	err = csv.SaveCSV(dir+"/"+relMDManuscript, [][]string{
		{"key", "elements_books", "additional_content", "elements_content"},
	})
	if err != nil {
		t.Fatalf("save manuscript metadata csv: %v", err)
	}

	store := NewEditionCSV(dir, nil)
	err = store.upsertManuscript(&model.Edition{
		Key:                   "ms_60",
		IsManuscript:          true,
		IsElements:            true,
		Books:                 []int{1, 2, 3, 5},
		AdditionalContent:     []string{"copy of Naples V A 13", "fragment"},
		ManuscriptElementsContent: strPtr("I-XV; abbreviated version"),
		ManuscriptClass:       "Latin Boethius manuscripts",
		ManuscriptSubclass:    strPtr("Mc"),
		ShortTitle:            "Test manuscript",
		ManuscriptYearFrom:    intPtr(1500),
		ManuscriptYearTo:      intPtr(1600),
		Title:                 strPtr("Long title"),
		ShortTitleSource:      "source",
	})
	if err != nil {
		t.Fatalf("upsert manuscript: %v", err)
	}

	_, rows, err := csv.LoadCSVRecords(dir + "/" + relMDManuscript)
	if err != nil {
		t.Fatalf("load manuscript metadata csv: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 manuscript metadata row, got %d", len(rows))
	}
	if rows[0]["elements_books"] != "1-3, 5" || rows[0]["additional_content"] != "copy of Naples V A 13, fragment" || rows[0]["elements_content"] != "I-XV; abbreviated version" {
		t.Fatalf("unexpected manuscript metadata row: %#v", rows[0])
	}
}

func strPtr(s string) *string { return &s }

func intPtr(i int) *int { return &i }

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
