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
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestGetAvailableCategoriesBulkDeduplicatesAndSortsCategories(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
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
		Meta:        common.Meta{ID: "ds_categories", Name: "Categories"},
		FacsimileID: "facsimile",
		DPI:         300,
	}))

	insertAnnotation := func(id string, categories ...string) {
		ann := &annotation.Annotation{
			Meta:      common.Meta{ID: id, Name: id},
			DatasetID: "ds_categories",
			Pages:     "1",
			Segmented: true,
		}
		require.NoError(t, annotationStore.InsertAnnotation(ann))
		altoDir := fileSysMgt.DatasetAnnotationAltoDir(ann)
		require.NoError(t, os.MkdirAll(altoDir, 0o755))
		tags := make([]alto.OtherTag, 0, len(categories))
		for i, category := range categories {
			tags = append(tags, alto.OtherTag{ID: string(rune('a' + i)), Label: category})
		}
		require.NoError(t, alto.SaveToFile(&alto.Alto{
			Tags: alto.Tags{OtherTags: tags},
		}, filepath.Join(altoDir, "page-0001.xml")))
	}

	insertAnnotation("ann_one", "MainZone", "Title")
	insertAnnotation("ann_two", "Commentary", "MainZone")

	datasetSvc := NewDatasetService(nil, nil, nil, datasetStore, fileSysMgt, nil, "", 1, 0)
	annotationSvc := NewAnnotationsService(datasetSvc, nil, nil, nil, fileSysMgt, annotationStore)
	categories, err := annotationSvc.GetAvailableCategoriesBulk([]annotation.Reference{
		{DatasetID: "ds_categories", ID: "ann_one"},
		{DatasetID: "ds_categories", ID: "ann_two"},
		{DatasetID: "ds_categories", ID: "ann_one"},
		{},
	})

	require.NoError(t, err)
	require.Equal(t, []string{"Commentary", "MainZone", "Title"}, categories)
}
