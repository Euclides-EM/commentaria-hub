package normalize

import "testing"

func TestLanguage(t *testing.T) {

	tests := []struct {
		input string
		want  string
	}{
		{"latin", "Latin"},
		{"greek", "Greek"},
		{"françois", "French"},
		{"italien", "Italian"},
		{"spanish", "Spanish"},
		{"german", "German"},
		{"nederduyts", "Dutch"},
		{"arabic", "Arabic"},
		{"english", "English"},
		{"romance", "General-Vernacular"},

		// Multiple languages
		{"latin, greek", "Latin::Greek"},
		{"aristotele alijsque græcis & latinis autoribus", "Latin::Greek"},
	}

	for _, tt := range tests {
		if got := Language(tt.input); got != tt.want {
			t.Errorf("Language(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
