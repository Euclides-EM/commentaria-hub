package store

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/samber/lo"
)

// Dataset storage layout:
//
// <baseDir>/
// └─ <dataset_id>/
//    ├─ <edition_id>_<facsimile_id>.pdf
//    ├─ imgs/
//    │  ├─ page-0001.png
//    │  ├─ page-0002.png
//    │  └─ ...
//    └─ annotations/
//       └─ <annotation_id>/
//          ├─ alto/
//          │  ├─ page-0001.xml
//          │  ├─ page-0002.xml
//          │  └─ ...
//          └─ yolo/
//             ├─ images/
//             │  ├─ page-0001.jpg
//             │  ├─ page-0002.jpg
//             │  └─ ...
//             ├─ labels/
//             │  ├─ page-0001.txt
//             │  ├─ page-0002.txt
//             │  └─ ...
//             ├─ config.yml
//             └─ labelmap.txt

type FileSystemManager struct {
	baseDir     string
	trainingDir string
	modelsDir   string
}

func NewFileSystemManager(baseDir, trainingDir, modelsDir string) *FileSystemManager {
	return &FileSystemManager{
		baseDir:     baseDir,
		trainingDir: trainingDir,
		modelsDir:   modelsDir,
	}
}

func (m *FileSystemManager) DatasetPDFPath(ds *model.Dataset) string {
	return path.Join(m.baseDir, ds.ID, fmt.Sprintf("%s_%s.pdf", ds.EditionID, ds.FacsimileID))
}

func (m *FileSystemManager) DatasetImagesDir(ds *model.Dataset) string {
	return path.Join(m.baseDir, ds.ID, "imgs")
}

func (m *FileSystemManager) datasetAnnotationsPath(ann *model.Annotation) string {
	return path.Join(m.baseDir, ann.DatasetID, "annotations", ann.ID)
}

func (m *FileSystemManager) DatasetAnnotationAltoDir(ann *model.Annotation) string {
	return path.Join(m.datasetAnnotationsPath(ann), "alto")
}

func (m *FileSystemManager) DatasetAnnotationYoloDir(ann *model.Annotation) string {
	return path.Join(m.datasetAnnotationsPath(ann), "yolo")
}

func (m *FileSystemManager) ModelPath(model *model.Model) string {
	return path.Join(m.modelsDir, model.LocalPath)
}

func (m *FileSystemManager) TrainingDir(t *model.Training) string {
	return path.Join(m.trainingDir, t.ID)
}

func (m *FileSystemManager) CleanupLocalStore(dryRun bool, annsMap map[string][]*model.Annotation, dss []*model.Dataset) ([]string, error) {
	var toDelete []string

	dataDir, err := filepath.Abs(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("could not get abs path for data dir: %v", err)
	}

	dsIDs := make([]string, 0)
	for _, ds := range dss {
		dsIDs = append(dsIDs, ds.ID)
	}

	des, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, de := range des {
		if !de.IsDir() {
			p := filepath.Join(dataDir, de.Name())
			toDelete = append(toDelete, p)
		}
		dsID := de.Name()
		if !slices.Contains(dsIDs, dsID) {
			p := filepath.Join(dataDir, dsID)
			toDelete = append(toDelete, p)
			continue
		}

		ddes, err := os.ReadDir(filepath.Join(dataDir, dsID))
		if err != nil {
			return nil, fmt.Errorf("cannot read dataset dir %s: %w", dsID, err)
		}
		for _, dde := range ddes {
			if filepath.Ext(dde.Name()) == ".pdf" {
				continue
			}
			if !dde.IsDir() {
				toDelete = append(toDelete, dde.Name())
				continue
			}
			if !slices.Contains([]string{"imgs", "annotations"}, dde.Name()) {
				p := filepath.Join(dataDir, dsID, dde.Name())
				toDelete = append(toDelete, p)
			}
		}

		annsIDs := lo.Map(annsMap[dsID], func(ann *model.Annotation, _ int) string {
			return ann.ID
		})

		annotationsPath := filepath.Join(dataDir, dsID, "annotations")
		ddes, err = os.ReadDir(annotationsPath)
		if err != nil {
			return nil, fmt.Errorf("cannot read dataset dir %s: %w", dsID, err)
		}

		for _, dde := range ddes {
			if !dde.IsDir() {
				p := filepath.Join(annotationsPath, dde.Name())
				toDelete = append(toDelete, p)
				continue
			}
			annID := dde.Name()
			found := false
			for _, existingAnnID := range annsIDs {
				if annID == existingAnnID {
					found = true
					break
				}
			}
			if !found {
				p := filepath.Join(annotationsPath, annID)
				toDelete = append(toDelete, p)
			}
		}
	}

	if !dryRun {
		for _, p := range toDelete {
			log.Printf("Deleting %s", p)
			if err = os.RemoveAll(p); err != nil {
				return nil, fmt.Errorf("failed to delete %s: %w", p, err)
			}
		}
	}

	return toDelete, nil
}
