package store

import (
	"testing"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/csv"
)

func TestLoadEditionByKeyKeepsManuscriptElementsBooksSeparateFromBooks(t *testing.T) {
	dir := t.TempDir()

	err := csv.SaveCSV(dir+"/"+relItemsManuscript, [][]string{
		{"key", "class", "subclass", "repository", "city", "languages", "compositors", "long_title", "short_title", "short_title_source", "year_from", "year_to", "year_is_approximate", "notes", "has_diagrams"},
		{"ms_1", "Latin manuscripts", "Subclass A", "", "", "", "", "", "Test manuscript", "source", "1200", "1250", "True", "", ""},
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
		{"key", "class", "subclass", "repository", "city", "languages", "compositors", "long_title", "short_title", "short_title_source", "year_from", "year_to", "year_is_approximate", "notes", "has_diagrams"},
		{"ms_60", "Latin Boethius manuscripts", "Mc", "", "", "", "", "", "", "", "1500", "1600", "True", "", ""},
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
