package search

import "testing"

type testEdition struct {
	ShortTitle string  `json:"shortTitle"`
	Title      *string `json:"title"`
	TitleEN    *string `json:"title_EN"`
	Year       *string `json:"year"`
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

func floatPtr(v float64) *float64 {
	return &v
}
