package tei

import "testing"

func TestMarkdownTableLineToTextPreservesEmptyCells(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "leading and trailing empty cells",
			line: "|  | □ en | □ en |  |",
			want: "\t□ en\t□ en\t",
		},
		{
			name: "leading empty cell",
			line: "|  | □ el (?) | □ nc + □ en | □ ec n. 47. I. |",
			want: "\t□ el (?)\t□ nc + □ en\t□ ec n. 47. I.",
		},
		{
			name: "interior empty cell",
			line: "| α. | ▭ alc + □ el |  | □ ec n. 3. Gr. I. |",
			want: "α.\t▭ alc + □ el\t\t□ ec n. 3. Gr. I.",
		},
		{
			name: "surrounding whitespace",
			line: "  |  | eh | Senckstr. | n. 3. Vorb. |  ",
			want: "\teh\tSenckstr.\tn. 3. Vorb.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := markdownTableLineToText(tt.line); got != tt.want {
				t.Fatalf("markdownTableLineToText() = %q, want %q", got, tt.want)
			}
		})
	}
}
