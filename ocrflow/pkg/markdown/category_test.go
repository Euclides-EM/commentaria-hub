package markdown

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractCategoryContentsExpandsDropcaps(t *testing.T) {
	md := &Markdown{Content: `## 1 {dropcap:P|lines=8|style=decorated|decoration="foliate ornamental initial in a square frame"}Vnctum, eſt quod partes non habet.`}

	categories, err := ExtractCategoryContentsFromMarkdown(md, nil, " / ", true)
	require.NoError(t, err)
	require.Equal(t, []Category{{
		Category: "header2",
		Content:  "1 PVnctum, eſt quod partes non habet.",
	}}, categories)
}

func TestExtractCategoryContentsIncludesCuratedHeadingsByDefault(t *testing.T) {
	md := &Markdown{Content: "[Curated heading level=1: Dedications]\n\n## Printed heading\n"}

	categories, err := ExtractCategoryContentsFromMarkdown(md, nil, " / ", true)
	require.NoError(t, err)
	require.Equal(t, []Category{
		{Category: "curated-heading1", Content: "Dedications"},
		{Category: "header2", Content: "Printed heading"},
	}, categories)
}

func TestExtractCategoryContentsCanExcludeCuratedHeadings(t *testing.T) {
	md := &Markdown{Content: "[Curated heading level=1: Dedications]\n\n## Printed heading\n"}

	categories, err := ExtractCategoryContentsFromMarkdown(md, nil, " / ", false)
	require.NoError(t, err)
	require.Equal(t, []Category{{Category: "header2", Content: "Printed heading"}}, categories)
}

func TestParseCuratedHeadingRequiresCanonicalSyntax(t *testing.T) {
	level, content := ParseCuratedHeading("[Curated heading level=12: Editorial structure]")
	require.Equal(t, 12, level)
	require.Equal(t, "Editorial structure", content)

	for _, invalid := range []string{
		"[Curated heading level=0: Invalid]",
		"[curated heading level=1: Invalid]",
		"[Curated heading level=1:]",
	} {
		level, content = ParseCuratedHeading(invalid)
		require.Zero(t, level, invalid)
		require.Empty(t, content, invalid)
	}
}
