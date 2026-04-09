package search

import (
	"slices"
	"testing"
)

type testEdition struct {
	ShortTitle          string  `json:"shortTitle"`
	Title               *string `json:"title"`
	TitleEN             *string `json:"title_EN"`
	Year                *string `json:"year"`
	IsManuscript        bool    `json:"isManuscript"`
	ManuscriptYearFrom  *int    `json:"manuscriptYearFrom"`
	ManuscriptYearTo    *int    `json:"manuscriptYearTo"`
}

func TestFilterFuncAppliesTextSearchWhenNonStrictRangeFieldIsMissing(t *testing.T) {
	query := Query{
		TextSearch:       "ttttttt",
		TextSearchFields: []string{"shortTitle", "title", "title_EN"},
		RangeFilter: map[string]Range{
			"year": {
				Min:    floatPtr(1482),
				Max:    floatPtr(1855),
				Strict: false,
			},
		},
	}

	item := testEdition{
		ShortTitle: "Euclid",
	}

	if query.FilterFunc()(item) {
		t.Fatalf("expected item without matching text to be filtered out")
	}
}

func TestFilterFuncAllowsMatchingTextWhenNonStrictRangeFieldIsMissing(t *testing.T) {
	query := Query{
		TextSearch:       "euclid",
		TextSearchFields: []string{"shortTitle", "title", "title_EN"},
		RangeFilter: map[string]Range{
			"year": {
				Min:    floatPtr(1482),
				Max:    floatPtr(1855),
				Strict: false,
			},
		},
	}

	item := testEdition{
		ShortTitle: "Euclid",
	}

	if !query.FilterFunc()(item) {
		t.Fatalf("expected item with matching text to pass")
	}
}

func TestFilterFuncUsesManuscriptYearRangeForYearFilter(t *testing.T) {
	query := Query{
		RangeFilter: map[string]Range{
			"year": {
				Min: floatPtr(1200),
				Max: floatPtr(1300),
			},
		},
	}

	item := testEdition{
		IsManuscript:       true,
		ManuscriptYearFrom: intPtr(1250),
		ManuscriptYearTo:   intPtr(1350),
	}

	if !query.FilterFunc()(item) {
		t.Fatalf("expected manuscript date span overlapping the filter to pass")
	}
}

func TestFilterFuncRejectsManuscriptYearRangeOutsideYearFilter(t *testing.T) {
	query := Query{
		RangeFilter: map[string]Range{
			"year": {
				Min: floatPtr(1200),
				Max: floatPtr(1300),
			},
		},
	}

	item := testEdition{
		IsManuscript:       true,
		ManuscriptYearFrom: intPtr(1351),
		ManuscriptYearTo:   intPtr(1400),
	}

	if query.FilterFunc()(item) {
		t.Fatalf("expected manuscript date span outside the filter to be rejected")
	}
}

func TestFilterFuncTreatsUndatedManuscriptLikeMissingPrintYearInNonStrictMode(t *testing.T) {
	query := Query{
		RangeFilter: map[string]Range{
			"year": {
				Min:    floatPtr(1200),
				Max:    floatPtr(1300),
				Strict: false,
			},
		},
	}

	item := testEdition{
		IsManuscript: true,
	}

	if !query.FilterFunc()(item) {
		t.Fatalf("expected undated manuscript to pass non-strict year filtering")
	}
}

func TestFilterFuncRejectsUndatedManuscriptInStrictMode(t *testing.T) {
	query := Query{
		RangeFilter: map[string]Range{
			"year": {
				Min:    floatPtr(1200),
				Max:    floatPtr(1300),
				Strict: true,
			},
		},
	}

	item := testEdition{
		IsManuscript: true,
	}

	if query.FilterFunc()(item) {
		t.Fatalf("expected undated manuscript to fail strict year filtering")
	}
}

func TestOrderByYearUsesManuscriptYears(t *testing.T) {
	query := Query{
		OrderBy: []OrderByOption{
			{Field: "year"},
		},
	}

	items := []testEdition{
		{
			ShortTitle:         "print",
			Year:               strPtr("1500"),
		},
		{
			ShortTitle:         "manuscript",
			IsManuscript:       true,
			ManuscriptYearFrom: intPtr(1200),
			ManuscriptYearTo:   intPtr(1300),
		},
	}

	slices.SortFunc(items, query.OrderByFunc())

	if items[0].ShortTitle != "manuscript" {
		t.Fatalf("expected manuscript item to sort by manuscript year, got first item %q", items[0].ShortTitle)
	}
}

func TestOrderByYearUsesManuscriptYearToWhenYearFromMissing(t *testing.T) {
	query := Query{
		OrderBy: []OrderByOption{
			{Field: "year"},
		},
	}

	items := []testEdition{
		{
			ShortTitle:         "later",
			IsManuscript:       true,
			ManuscriptYearTo:   intPtr(1400),
		},
		{
			ShortTitle:         "earlier",
			Year:               strPtr("1300"),
		},
	}

	slices.SortFunc(items, query.OrderByFunc())

	if items[0].ShortTitle != "earlier" {
		t.Fatalf("expected manuscript fallback year_to to participate in ordering, got first item %q", items[0].ShortTitle)
	}
}

func TestOrderByYearTreatsUndatedManuscriptLikeMissingYear(t *testing.T) {
	query := Query{
		OrderBy: []OrderByOption{
			{Field: "year"},
		},
	}

	items := []testEdition{
		{
			ShortTitle:   "dated",
			Year:         strPtr("1500"),
		},
		{
			ShortTitle:   "undated manuscript",
			IsManuscript: true,
		},
	}

	slices.SortFunc(items, query.OrderByFunc())

	if items[1].ShortTitle != "undated manuscript" {
		t.Fatalf("expected undated manuscript to sort last, got last item %q", items[1].ShortTitle)
	}
}

func TestOrderByYearTreatsUndatedManuscriptAsLastInDescendingOrder(t *testing.T) {
	query := Query{
		OrderBy: []OrderByOption{
			{Field: "year", Descending: true},
		},
	}

	items := []testEdition{
		{
			ShortTitle: "dated",
			Year:       strPtr("1500"),
		},
		{
			ShortTitle:   "undated manuscript",
			IsManuscript: true,
		},
	}

	slices.SortFunc(items, query.OrderByFunc())

	if items[1].ShortTitle != "undated manuscript" {
		t.Fatalf("expected undated manuscript to sort last in descending order, got last item %q", items[1].ShortTitle)
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func strPtr(v string) *string {
	return &v
}
