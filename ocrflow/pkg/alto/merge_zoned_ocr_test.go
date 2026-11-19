package alto

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeZonedOCRClipsSplitsAndRemovesLines(t *testing.T) {
	segmentation := testSegmentationALTO()
	ocr := testOCRALTO()

	stats, err := MergeZonedOCR(segmentation, ocr, []string{"MainZone"}, []string{"GraphicZone-Diagram"})
	require.NoError(t, err)
	require.Equal(t, 2, stats.InputLines)
	require.Equal(t, 2, stats.OutputLines)
	require.Equal(t, 1, stats.RemovedLines)
	require.Equal(t, 1, stats.SplitLines)
	require.Equal(t, 12, stats.OutputRunes)
	require.Equal(t, 20, stats.RemovedRunes) // 8 clipped runes + 12 from the outside line.

	main := segmentation.Layout.Page[0].PrintSpace.TextBlocks[0]
	ignored := segmentation.Layout.Page[0].PrintSpace.TextBlocks[1]
	require.Len(t, main.Lines, 2)
	require.Empty(t, ignored.Lines)
	require.Equal(t, "234567", main.Lines[0].Strings[0].Content)
	require.Equal(t, "cdefgh", main.Lines[1].Strings[0].Content)
	require.InDelta(t, 100, main.Lines[0].HPOS, 0.001)
	require.InDelta(t, 300, main.Lines[0].Width, 0.001)
	require.InDelta(t, 600, main.Lines[1].HPOS, 0.001)
	require.InDelta(t, 300, main.Lines[1].Width, 0.001)
	require.Equal(t, "line-1__part_1", main.Lines[0].ID)
	require.Equal(t, "line-1__part_2", main.Lines[1].ID)
	require.NotContains(t, main.Lines[0].Shape.Polygon.Points, "600 60")
}

func TestMergeZonedOCRDirsMatchesNestedOCRPages(t *testing.T) {
	root := t.TempDir()
	segDir := filepath.Join(root, "segmentation")
	ocrDir := filepath.Join(root, "ocr")
	outDir := filepath.Join(root, "merged")
	require.NoError(t, os.MkdirAll(segDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(ocrDir, "page-0001"), 0o755))
	require.NoError(t, SaveToFile(testSegmentationALTO(), filepath.Join(segDir, "page-0001.xml")))
	require.NoError(t, SaveToFile(testOCRALTO(), filepath.Join(ocrDir, "page-0001", "original.xml")))

	stats, err := MergeZonedOCRDirs(segDir, ocrDir, outDir, []string{"MainZone"}, []string{"GraphicZone-Diagram"})
	require.NoError(t, err)
	require.Equal(t, 1, stats.Pages)
	mergedPath := filepath.Join(outDir, "page-0001", "original.xml")
	merged, err := LoadFromFile(mergedPath)
	require.NoError(t, err)
	require.Len(t, merged.Layout.Page[0].PrintSpace.TextBlocks[0].Lines, 2)
	_, err = os.Stat(filepath.Join(outDir, "page-0001.xml"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestMergeZonedOCRDirsAppliesOrderedReassignments(t *testing.T) {
	root := t.TempDir()
	segDir := filepath.Join(root, "segmentation")
	ocrDir := filepath.Join(root, "ocr")
	outDir := filepath.Join(root, "merged")
	require.NoError(t, os.MkdirAll(segDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(ocrDir, "page-0001"), 0o755))
	require.NoError(t, SaveToFile(testSegmentationALTO(), filepath.Join(segDir, "page-0001.xml")))
	require.NoError(t, SaveToFile(testOCRALTO(), filepath.Join(ocrDir, "page-0001", "original.xml")))

	stats, err := MergeZonedOCRDirsWithReassignments(
		segDir, ocrDir, outDir,
		[]string{"MainZone"}, []string{"GraphicZone-Diagram"},
		[]LineReassignment{
			{FromCategory: "MainZone", ToCategory: "MainZone-Head--Section", PrecisionPx: 5, MinOverlap: 0.85},
			{FromCategory: "MainZone", ToCategory: "MainZone-Head--Book", PrecisionPx: 5, MinOverlap: 0.6},
		},
	)
	require.NoError(t, err)
	require.Equal(t, 2, stats.ReassignedLines)

	merged, err := LoadFromFile(filepath.Join(outDir, "page-0001", "original.xml"))
	require.NoError(t, err)
	blocks := merged.Layout.Page[0].PrintSpace.TextBlocks
	require.Empty(t, blocks[0].Lines)
	require.Len(t, blocks[2].Lines, 1)
	require.Equal(t, "line-1__part_1", blocks[2].Lines[0].ID)
	require.Len(t, blocks[3].Lines, 1)
	require.Equal(t, "line-1__part_2", blocks[3].Lines[0].ID)
}

func TestMergeZonedOCRUsesExactCategoryLabels(t *testing.T) {
	segmentation := testSegmentationALTO()
	segmentation.Tags.OtherTags[0].Label = "MainZone-Head--Book"

	stats, err := MergeZonedOCR(segmentation, testOCRALTO(), []string{"MainZone"}, nil)
	require.NoError(t, err)
	require.Zero(t, stats.OutputLines)
	require.Empty(t, segmentation.Layout.Page[0].PrintSpace.TextBlocks[0].Lines)
}

func testSegmentationALTO() *Alto {
	return &Alto{
		Tags: Tags{OtherTags: []OtherTag{
			{ID: "REGION_MAIN", Label: "MainZone"},
			{ID: "REGION_IGNORE", Label: "GraphicZone-Diagram"},
			{ID: "REGION_SECTION", Label: "MainZone-Head--Section"},
			{ID: "REGION_BOOK", Label: "MainZone-Head--Book"},
		}},
		Layout: Layout{Page: []Page{{
			Width: 1000, Height: 100, ID: "page-0001",
			PrintSpace: PrintSpace{Width: 1000, Height: 100, TextBlocks: []TextBlock{
				{ID: "main", TagRefs: "REGION_MAIN", HPOS: 100, VPOS: 20, Width: 800, Height: 60,
					Shape: Shape{Polygon: Polygon{Points: "100 20 900 20 900 80 100 80 100 20"}},
					Lines: []TextLine{{ID: "old-segmentation-line"}}},
				{ID: "diagram", TagRefs: "REGION_IGNORE", HPOS: 400, VPOS: 20, Width: 200, Height: 60,
					Shape: Shape{Polygon: Polygon{Points: "400 20 600 20 600 80 400 80 400 20"}},
					Lines: []TextLine{{ID: "old-ignore-line"}}},
				{ID: "section", TagRefs: "REGION_SECTION", HPOS: 100, VPOS: 20, Width: 300, Height: 60,
					Shape: Shape{Polygon: Polygon{Points: "100 20 400 20 400 80 100 80 100 20"}}},
				{ID: "book", TagRefs: "REGION_BOOK", HPOS: 600, VPOS: 20, Width: 180, Height: 60,
					Shape: Shape{Polygon: Polygon{Points: "600 20 780 20 780 80 600 80 600 20"}}},
			}},
		}}},
	}
}

func testOCRALTO() *Alto {
	return &Alto{
		Layout: Layout{Page: []Page{{
			Width: 100, Height: 100, ID: "page-0001",
			PrintSpace: PrintSpace{Width: 100, Height: 100, TextBlocks: []TextBlock{{
				ID: "ocr", Lines: []TextLine{
					{ID: "line-1", HPOS: 0, VPOS: 40, Width: 100, Height: 20, Baseline: "0 50 100 50",
						Shape:   Shape{Polygon: Polygon{Points: "0 40 100 40 100 60 0 60 0 40"}},
						Strings: []String{{Content: "0123456789abcdefghij", HPOS: 0, VPOS: 40, Width: 100, Height: 20}}},
					{ID: "outside", HPOS: 0, VPOS: 0, Width: 100, Height: 10, Baseline: "0 10 100 10",
						Shape:   Shape{Polygon: Polygon{Points: "0 0 100 0 100 10 0 10 0 0"}},
						Strings: []String{{Content: "outside text", HPOS: 0, VPOS: 0, Width: 100, Height: 10}}},
				},
			}}},
		}}},
	}
}
