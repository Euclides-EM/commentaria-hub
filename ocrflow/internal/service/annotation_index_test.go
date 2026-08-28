package service

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	_ "github.com/mattn/go-sqlite3"
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

func TestGetAnnotationIndexReturnsEmptyForUnpreparedAnnotation(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	schema, err := os.ReadFile(filepath.Join("..", "migrations", "ocrflow", "1774207510_tables.sql"))
	require.NoError(t, err)
	_, err = db.Exec(string(schema))
	require.NoError(t, err)

	fileSysMgt := filesys.NewFileSystemManager(t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir())
	datasetStore := store.NewDatasetSQL(db, fileSysMgt)
	annotationStore := store.NewAnnotationSQL(db)
	require.NoError(t, datasetStore.InsertDataset(&model.Dataset{
		Meta:        common.Meta{ID: "ds_unprepared", Name: "Unprepared"},
		EditionID:   "XANRSM",
		FacsimileID: "facsimile",
		DPI:         300,
	}))
	require.NoError(t, annotationStore.InsertAnnotation(&annotation.Annotation{
		Meta:      common.Meta{ID: "ann_unprepared", Name: "Unprepared"},
		DatasetID: "ds_unprepared",
		Pages:     "1",
		Segmented: false,
		Ocred:     false,
	}))

	datasetSvc := NewDatasetService(nil, nil, nil, datasetStore, fileSysMgt, "", 1, 0)
	annotationSvc := NewAnnotationsService(datasetSvc, nil, nil, nil, fileSysMgt, annotationStore)

	index, err := annotationSvc.GetAnnotationIndex("ds_unprepared", "ann_unprepared", nil)
	require.NoError(t, err)
	require.Equal(t, "ds_unprepared", index.DatasetID)
	require.Equal(t, "ann_unprepared", index.AnnotationID)
	require.Empty(t, index.Nodes)
}

func TestGetAnnotationIndexPrefersAnnotationMarkdownOverEditionMarkdown(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	schema, err := os.ReadFile(filepath.Join("..", "migrations", "ocrflow", "1774207510_tables.sql"))
	require.NoError(t, err)
	_, err = db.Exec(string(schema))
	require.NoError(t, err)

	baseDir := t.TempDir()
	fileSysMgt := filesys.NewFileSystemManager(baseDir, t.TempDir(), t.TempDir(), t.TempDir())
	datasetStore := store.NewDatasetSQL(db, fileSysMgt)
	annotationStore := store.NewAnnotationSQL(db)
	require.NoError(t, datasetStore.InsertDataset(&model.Dataset{
		Meta:        common.Meta{ID: "ds_priority", Name: "Priority"},
		EditionID:   "edition_priority",
		FacsimileID: "facsimile",
		DPI:         300,
	}))
	ann := &annotation.Annotation{
		Meta:      common.Meta{ID: "ann_priority", Name: "Priority"},
		DatasetID: "ds_priority",
		Pages:     "1",
		Segmented: false,
		Ocred:     true,
	}
	require.NoError(t, annotationStore.InsertAnnotation(ann))

	writeMarkdownPage := func(dir, content string) {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "page-0001"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "page-0001", "original.md"), []byte(content), 0o644))
	}
	writeMarkdownPage(filepath.Join(baseDir, "ds_priority", "annotations", "ann_priority", "transcriptions"), "# Annotation wins\n")
	writeMarkdownPage(filepath.Join(baseDir, "transcriptions", "edition_priority"), "# Edition loses\n")

	datasetSvc := NewDatasetService(nil, nil, nil, datasetStore, fileSysMgt, "", 1, 0)
	annotationSvc := NewAnnotationsService(datasetSvc, nil, nil, nil, fileSysMgt, annotationStore)

	index, err := annotationSvc.GetAnnotationIndex("ds_priority", "ann_priority", nil)
	require.NoError(t, err)
	require.Len(t, index.Nodes, 1)
	require.Equal(t, "Annotation wins", index.Nodes[0].Content)
	require.Equal(t, "header1", index.Nodes[0].Category)
}
