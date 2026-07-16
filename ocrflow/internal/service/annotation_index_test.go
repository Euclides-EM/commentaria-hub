package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderedMarkdownHeaderCategoriesSortsByHeaderLevel(t *testing.T) {
	got := orderedMarkdownHeaderCategories(map[string]struct{}{
		"header3": {},
		"header1": {},
		"header2": {},
	})

	require.Equal(t, []string{"header1", "header2", "header3"}, got)
}

func TestBuildNodesNestsMarkdownHeadersInDocumentOrder(t *testing.T) {
	nodes := buildNodes([]string{"header1", "header2", "header3"}, []categoryPageContent{
		{page: 1, category: "header1", content: "Book I"},
		{page: 1, category: "header2", content: "Definitions"},
		{page: 2, category: "header3", content: "I."},
		{page: 3, category: "header3", content: "II."},
		{page: 4, category: "header2", content: "Propositions"},
		{page: 5, category: "header3", content: "I."},
	})

	require.Len(t, nodes, 1)
	require.Equal(t, "Book I", nodes[0].Content)
	require.Len(t, nodes[0].Children, 2)
	require.Equal(t, "Definitions", nodes[0].Children[0].Content)
	require.Len(t, nodes[0].Children[0].Children, 2)
	require.Equal(t, "I.", nodes[0].Children[0].Children[0].Content)
	require.Equal(t, "II.", nodes[0].Children[0].Children[1].Content)
	require.Equal(t, "Propositions", nodes[0].Children[1].Content)
	require.Len(t, nodes[0].Children[1].Children, 1)
	require.Equal(t, "I.", nodes[0].Children[1].Children[0].Content)
}

func TestBuildNodesRetainsHeaderWhenIntermediateLevelIsMissing(t *testing.T) {
	nodes := buildNodes([]string{"header1", "header2", "header3", "header4"}, []categoryPageContent{
		{page: 18, category: "header1", content: "Book I"},
		{page: 18, category: "header2", content: "Definitions"},
		{page: 18, category: "header4", content: "I."},
		{page: 19, category: "header3", content: "II."},
	})

	require.Len(t, nodes, 1)
	require.Len(t, nodes[0].Children, 1)
	definitions := nodes[0].Children[0]
	require.Len(t, definitions.Children, 2)
	require.Equal(t, "header4", definitions.Children[0].Category)
	require.Equal(t, "I.", definitions.Children[0].Content)
	require.Equal(t, "18", definitions.Children[0].Location.Page)
	require.Equal(t, "header3", definitions.Children[1].Category)
	require.Equal(t, "II.", definitions.Children[1].Content)
	require.Equal(t, "19", definitions.Children[1].Location.Page)
}
