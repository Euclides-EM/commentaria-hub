package textmatch

import "testing"

func TestFindLoosePhraseMatches(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		featureValue string
		want         []string
	}{
		{
			name:         "matches exact phrase",
			text:         "And some of EUCLID's Demonstrations are Restored.",
			featureValue: "EUCLID's",
			want:         []string{"EUCLID's"},
		},
		{
			name:         "matches across hyphen and newline",
			text:         "EVCLI-\nDIS",
			featureValue: "EVCLIDIS",
			want:         []string{"EVCLI-\nDIS"},
		},
		{
			name:         "matches with spaces removed from feature value",
			text:         "And some of EUCLID's Demonstrations are Restored.",
			featureValue: "EUCLID ' s",
			want:         []string{"EUCLID's"},
		},
		{
			name:         "matches multiple overlapping occurrences",
			text:         "ABABA",
			featureValue: "ABA",
			want:         []string{"ABA", "ABA"},
		},
		{
			name:         "empty normalized feature value returns no matches",
			text:         "ABC",
			featureValue: " - \n ",
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindLoosePhraseMatches(tt.text, tt.featureValue)
			if len(got) != len(tt.want) {
				t.Fatalf("FindLoosePhraseMatches() len = %d, want %d", len(got), len(tt.want))
			}
			for i, span := range got {
				match := tt.text[span[0]:span[1]]
				if match != tt.want[i] {
					t.Fatalf("FindLoosePhraseMatches()[%d] = %q, want %q", i, match, tt.want[i])
				}
			}
		})
	}
}
