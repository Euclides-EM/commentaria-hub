package tei

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/markdown"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

func TestMarkdownTableLineToTextPreservesEmptyCells(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "leading and trailing empty cells",
			line: "|  | □ en | □ en |  |",
			want: " | □ en | □ en | ",
		},
		{
			name: "leading empty cell",
			line: "|  | □ el (?) | □ nc + □ en | □ ec n. 47. I. |",
			want: " | □ el (?) | □ nc + □ en | □ ec n. 47. I.",
		},
		{
			name: "interior empty cell",
			line: "| α. | ▭ alc + □ el |  | □ ec n. 3. Gr. I. |",
			want: "α. | ▭ alc + □ el |  | □ ec n. 3. Gr. I.",
		},
		{
			name: "surrounding whitespace",
			line: "  |  | eh | Senckstr. | n. 3. Vorb. |  ",
			want: " | eh | Senckstr. | n. 3. Vorb.",
		},
		{
			name: "escaped pipe is cell content",
			line: `| 40 | ag 2\|2 b | ag 2\|2 bc |`,
			want: "40 | ag 2|2 b | ag 2|2 bc",
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

func TestMarkdownBlocksUseCanonicalDialect(t *testing.T) {
	md := &markdown.Markdown{Content: `<!-- Running title: LIBER I. -->

# Book

[Margin]
printed note
[/Margin]

[Other type="binding"]
label
[/Other]

[Diagram: circle labelled A]

[Illustration]

[Calculation]
  12
+ 34
[/Calculation]

[Blank page]

| A | B |
|---|---|
| 2\|2 | x |
`}
	abs := markdownBlocksToABs("12", md)
	wantTypes := []string{
		"running-title", "header1", "margin", "other:binding", "diagram",
		"illustration", "calculation", "blank-page", "table",
	}
	if len(abs) != len(wantTypes) {
		t.Fatalf("got %d blocks, want %d: %#v", len(abs), len(wantTypes), abs)
	}
	for i, want := range wantTypes {
		if abs[i].Type != want {
			t.Errorf("block %d type = %q, want %q", i, abs[i].Type, want)
		}
	}
	if got := inlineText(abs[4].Lines[0].Nodes); got != "circle labelled A" {
		t.Errorf("diagram description = %q", got)
	}
	if got := inlineText(abs[8].Lines[1].Nodes); got != "2|2 | x" {
		t.Errorf("table row = %q", got)
	}
}

func TestMarkdownBlocksRejectLegacyObjectSyntax(t *testing.T) {
	abs := markdownBlocksToABs("1", &markdown.Markdown{Content: "*[Figure]*\n"})
	if len(abs) != 1 || abs[0].Type != "paragraph" {
		t.Fatalf("legacy object syntax was accepted: %#v", abs)
	}
}

func TestMarkdownInlineAnnotations(t *testing.T) {
	nodes := markdownInlineNodes(`{dropcap:P|lines=3|style=decorated|decoration="floral"}Rinted triangulun{printer-error-correction:triangulum} [illegible: 2 words] [unclear: AB]`)
	if got := inlineText(nodes); got != "PRinted triangulun [correction: triangulum] [illegible: 2 words] [unclear: AB]" {
		t.Fatalf("inline text = %q", got)
	}
	if nodes[0].Inline == nil || nodes[0].Inline.Rend != "dropcap lines=3 style=decorated decoration=floral" {
		t.Fatalf("dropcap node = %#v", nodes[0])
	}
}

func inlineText(nodes []model.ABNode) string {
	var result string
	for _, node := range nodes {
		if node.Inline != nil {
			result += node.Inline.Text
		} else {
			result += node.CharData
		}
	}
	return result
}
