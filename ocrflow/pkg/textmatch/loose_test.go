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
		{
			name:         "matches ' to ’ and other punctuation variations",
			text:         "THE\nELEMENTS\nOF\nEUCLID\nExplain’d,\nIn a New, but most Easie Method:\nTogether with The Use of every Proposition through\nall parts of the Mathematicks.\nWritten in French, by that Excellent\nMathematician,\nF. CLAUD. FRANCIS MILLIET de CHALES,\nof the Society of JESUS.\nNow made English, and a multitude of Errors Corrected, which\nhad escap’d in the Original.\nThe Third Edition.\nOXFORD,\nPrinted by L.L. for M. Gillyflower at the Spread-Eagle in West-\nminster-Hall, and W. Freeman at the Bible over-against the\nMiddle-Temple-Gate, in Fleet-street, 1700.",
			featureValue: "Explain'd",
			want:         []string{"Explain’d"},
		},
		{
			name:         "matches long phrase with various whitespace and punctuation",
			text:         "THE\nELEMENTS\nOF\nEUCLID\nExplain’d,\nIn a New, but most Easie Method:\nTogether with The Use of every Proposition through\nall parts of the Mathematicks.\nWritten in French, by that Excellent\nMathematician,\nF. CLAUD. FRANCIS MILLIET de CHALES,\nof the Society of JESUS.\nNow made English, and a multitude of Errors Corrected, which\nhad escap’d in the Original.\nThe Third Edition.\nOXFORD,\nPrinted by L.L. for M. Gillyflower at the Spread-Eagle in West-\nminster-Hall, and W. Freeman at the Bible over-against the\nMiddle-Temple-Gate, in Fleet-street, 1700.",
			featureValue: "Explain'd, In a New, but most Easie Method, Together with The Use of every Proposition through all parts of the Mathematicks",
			want:         []string{"Explain’d,\nIn a New, but most Easie Method:\nTogether with The Use of every Proposition through\nall parts of the Mathematicks"},
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
