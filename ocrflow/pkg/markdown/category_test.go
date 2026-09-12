package markdown

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractCategoryContentsExpandsDropcaps(t *testing.T) {
	md := &Markdown{Content: `## 1 {dropcap:P|lines=8|style=decorated|decoration="foliate ornamental initial in a square frame"}Vnctum, eſt quod partes non habet.`}

	categories, _, err := ExtractIndexContentsFromMarkdown(md, nil, nil)
	require.NoError(t, err)
	require.Equal(t, []Category{{
		Category:  "header2",
		Content:   "1 PVnctum, eſt quod partes non habet.",
		IndexType: DefaultIndexType,
	}}, categories)
}

func TestExtractCategoryContentsJoinsConsecutiveHeadersAtSameLevel(t *testing.T) {
	md := &Markdown{Content: `# NOVVEAVX ELEMENS DE GEOMETRIE.

# LIVRE PREMIER.

## Definitions

Body text.

## Propositions`}

	categories, _, err := ExtractIndexContentsFromMarkdown(md, nil, nil)
	require.NoError(t, err)
	require.Equal(t, []Category{
		{Category: "header1", Content: "NOVVEAVX ELEMENS DE GEOMETRIE. LIVRE PREMIER.", IndexType: DefaultIndexType},
		{Category: "header2", Content: "Definitions", IndexType: DefaultIndexType},
		{Category: "header2", Content: "Propositions", IndexType: DefaultIndexType},
	}, categories)
}

func TestParseHeaderBlockDoesNotJoinDifferentLevels(t *testing.T) {
	lines := []string{"# Book", "", "## Definitions"}

	level, content, next := ParseHeaderBlock(lines, 0)
	require.Equal(t, 1, level)
	require.Equal(t, "Book", content)
	require.Equal(t, 1, next)
}

func TestExtractCategoryContentsIncludesCuratedHeadingsByDefault(t *testing.T) {
	md := &Markdown{Content: "[Curated heading level=1: Dedications]\n\n## Printed heading\n"}

	categories, _, err := ExtractIndexContentsFromMarkdown(md, nil, nil)
	require.NoError(t, err)
	require.Equal(t, []Category{
		{Category: "curated-heading1", Content: "Dedications", IndexType: DefaultIndexType},
		{Category: "header2", Content: "Printed heading", IndexType: DefaultIndexType},
	}, categories)
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

func TestParseTypedCuratedHeading(t *testing.T) {
	heading, ok := ParseTypedCuratedHeading("[Curated heading level=2 type=myType: Parallel section]")
	require.True(t, ok)
	require.Equal(t, CuratedHeading{Level: 2, Type: "myType", Text: "Parallel section"}, heading)

	heading, ok = ParseTypedCuratedHeading("[Curated heading type=paragraph_order level=4: I.]")
	require.True(t, ok)
	require.Equal(t, CuratedHeading{Level: 4, Type: "paragraph_order", Text: "I."}, heading)

	heading, ok = ParseTypedCuratedHeading("[Curated heading level=1: Default section]")
	require.True(t, ok)
	require.Equal(t, DefaultIndexType, heading.Type)

	for _, invalid := range []string{
		"[Curated heading level=1 type=default: Reserved]",
		"[Curated heading level=1 type=my type: Invalid]",
		"[Curated heading level=1 type=\"quoted\": Invalid]",
		"[Curated heading type=myType: Missing level]",
		"[Curated heading level=01: Invalid level]",
		"[Curated heading level=1 level=2: Duplicate level]",
		"[Curated heading level=1 type=one type=two: Duplicate type]",
		"[Curated heading level=1 unknown=value: Invalid]",
	} {
		_, ok = ParseTypedCuratedHeading(invalid)
		require.False(t, ok, invalid)
	}
}

func TestExtractIndexContentsFiltersLayersAndReturnsAvailableTypes(t *testing.T) {
	md := &Markdown{Content: `# Printed book

[Curated heading level=2: Default division]

[Curated heading level=1 type=myType: Typed book]

[Curated heading type=myType2 level=2: Typed division]`}

	contents, availableTypes, err := ExtractIndexContentsFromMarkdown(
		md,
		nil,
		[]string{"default", "myType2"},
	)
	require.NoError(t, err)
	require.Equal(t, []string{"default", "myType", "myType2"}, availableTypes)
	require.Equal(t, []Category{
		{Category: "header1", Content: "Printed book", IndexType: "default"},
		{Category: "curated-heading2", Content: "Default division", IndexType: "default"},
		{Category: "curated-heading2", Content: "Typed division", IndexType: "myType2"},
	}, contents)
}
