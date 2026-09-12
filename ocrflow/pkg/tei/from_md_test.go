package tei

import (
	"strings"
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

[Curated heading level=2: Editorial division]

[Curated heading level=3 type=myType: Parallel division]

[Subhead]
Printed qualification of the heading
[/Subhead]

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
		"running-title", "header1", "curated-heading", "curated-heading", "subhead", "margin", "other:binding", "diagram",
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
	if abs[2].N != "2" || inlineText(abs[2].Lines[0].Nodes) != "Editorial division" {
		t.Errorf("curated heading = %#v", abs[2])
	}
	if abs[3].N != "3" || abs[3].Subtype != "myType" || inlineText(abs[3].Lines[0].Nodes) != "Parallel division" {
		t.Errorf("typed curated heading = %#v", abs[3])
	}
	if got := inlineText(abs[4].Lines[0].Nodes); got != "Printed qualification of the heading" {
		t.Errorf("subhead = %q", got)
	}
	if got := inlineText(abs[7].Lines[0].Nodes); got != "circle labelled A" {
		t.Errorf("diagram description = %q", got)
	}
	if got := inlineText(abs[11].Lines[1].Nodes); got != "2|2 | x" {
		t.Errorf("table row = %q", got)
	}
}

func TestMarkdownBlocksJoinConsecutiveHeadersAtSameLevel(t *testing.T) {
	abs := markdownBlocksToABs("39", &markdown.Markdown{Content: `# NOVVEAVX ELEMENS DE GEOMETRIE.

# LIVRE PREMIER.

Body text.`})

	if len(abs) != 2 {
		t.Fatalf("got %d blocks, want 2: %#v", len(abs), abs)
	}
	if abs[0].Type != "header1" {
		t.Fatalf("heading type = %q, want header1", abs[0].Type)
	}
	if got := inlineText(abs[0].Lines[0].Nodes); got != "NOVVEAVX ELEMENS DE GEOMETRIE. LIVRE PREMIER." {
		t.Fatalf("heading content = %q", got)
	}
}

func TestMarkdownSubheadBecomesTEIAB(t *testing.T) {
	doc, err := BuildTEIFromMarkdown("32", &markdown.Markdown{Content: `## Main heading

[Subhead]
Printed qualification.
[/Subhead]

Body text.`}, nil)
	if err != nil {
		t.Fatalf("BuildTEIFromMarkdown() error = %v", err)
	}

	xmlBytes, err := doc.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}
	xmlText := string(xmlBytes)
	if !strings.Contains(xmlText, `<ab xml:id="transcription_anon_blk_page_32_2" type="subhead">`) {
		t.Fatalf("TEI does not contain a subhead ab:\n%s", xmlText)
	}
	if !strings.Contains(xmlText, `>Printed qualification.</l>`) {
		t.Fatalf("TEI subhead content is not formatted correctly:\n%s", xmlText)
	}
	if strings.Contains(xmlText, "[Subhead]") || strings.Contains(xmlText, "[/Subhead]") {
		t.Fatalf("Markdown subhead markers leaked into TEI:\n%s", xmlText)
	}
}

func TestTypedCuratedHeadingPreservesLayerInTEI(t *testing.T) {
	doc, err := BuildTEIFromMarkdown("7", &markdown.Markdown{Content: "[Curated heading type=myType level=2: Parallel section]"}, nil)
	if err != nil {
		t.Fatalf("BuildTEIFromMarkdown() error = %v", err)
	}
	xmlBytes, err := doc.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}
	if !strings.Contains(string(xmlBytes), `type="curated-heading" subtype="myType" n="2"`) {
		t.Fatalf("typed curated heading attributes missing from TEI:\n%s", xmlBytes)
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

func TestMarkdownInlineEscapedAsteriskBecomesLiteralText(t *testing.T) {
	nodes := markdownInlineNodes(`\* 228. *Il manque une raye droite* de *c* à *d*.`)
	if got := inlineText(nodes); got != "* 228. Il manque une raye droite de c à d." {
		t.Fatalf("inline text = %q", got)
	}
	if len(nodes) != 7 {
		t.Fatalf("got %d nodes, want 7: %#v", len(nodes), nodes)
	}
	for _, index := range []int{1, 3, 5} {
		if nodes[index].Inline == nil || nodes[index].Inline.Rend != "italic" {
			t.Fatalf("node %d = %#v, want italic inline", index, nodes[index])
		}
	}
}

func TestMarkdownInlineEscapedAsteriskInMarginTEI(t *testing.T) {
	doc, err := BuildTEIFromMarkdown("38", &markdown.Markdown{Content: `[Margin]
\* Le chiffre eſt manqué, ne marque que [unclear: 260], & le ſuivant 221.
[/Margin]`}, nil)
	if err != nil {
		t.Fatalf("BuildTEIFromMarkdown() error = %v", err)
	}

	xmlBytes, err := doc.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}
	xmlText := string(xmlBytes)
	if !strings.Contains(xmlText, `>* Le chiffre eſt manqué, ne marque que `) {
		t.Fatalf("escaped asterisk was not rendered as literal text:\n%s", xmlText)
	}
	if strings.Contains(xmlText, `>\* Le chiffre`) {
		t.Fatalf("Markdown escape leaked into TEI:\n%s", xmlText)
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
