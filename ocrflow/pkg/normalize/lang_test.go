package normalize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguage(t *testing.T) {

	// todo: fix bug - the original returned by the func in actually the normalized string, not the original.
	//  once this is fixed, uncomment the test cases currently commented out.
	tests := []struct {
		input string
		want  []MappedOriginal
	}{
		{"latin", []MappedOriginal{{"latin", "Latin"}}},
		{"greek", []MappedOriginal{{"greek", "Greek"}}},
		// {"françois", []MappedOriginal{{"françois", "French"}}},
		{"italien", []MappedOriginal{{"italien", "Italian"}}},
		{"spanish", []MappedOriginal{{"spanish", "Spanish"}}},
		{"german", []MappedOriginal{{"german", "German"}}},
		{"nederduyts", []MappedOriginal{{"nederduyts", "Dutch"}}},
		{"arabic", []MappedOriginal{{"arabic", "Arabic"}}},
		{"english", []MappedOriginal{{"english", "English"}}},
		{"romance", []MappedOriginal{{"romance", "General-Vernacular"}}},

		// Multiple languages
		{"latin, greek", []MappedOriginal{{"latin", "Latin"}, {"greek", "Greek"}}},
		// {"aristotele alijsque græcis & latinis autoribus", []MappedOriginal{{"greek", "græcis"}, {"latin", "latinis"}}},
		// {"in Platone, Aristotele alijsque Græcis & Latinis autoribus", []MappedOriginal{{"Græcis", "Greek"}, {"Latinis", "Latin"}}},

		// No matches
		{"unknown language", []MappedOriginal{}},

		// Case insensitivity
		{"LATIN", []MappedOriginal{{"latin", "Latin"}}},

		// Leading/trailing whitespace
		{"  latin  ", []MappedOriginal{{"latin", "Latin"}}},

		// Empty input
		{"", []MappedOriginal{}},

		// Input with extra punctuation
		{"latin; greek", []MappedOriginal{{"latin", "Latin"}, {"greek", "Greek"}}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Language(tt.input)
			if !assert.ElementsMatch(t, got, tt.want) {
				t.Logf("Language(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
