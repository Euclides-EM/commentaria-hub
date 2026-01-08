package filesys

import (
	"fmt"
	"path"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
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

func (m *Manager) DatasetPDFPath(ds *model.Dataset) string {
	return path.Join(m.baseDir, ds.ID, fmt.Sprintf("%s_%s.pdf", ds.EditionID, ds.FacsimileID))
}

func (m *Manager) DatasetImagesDir(ds *model.Dataset) string {
	return path.Join(m.baseDir, ds.ID, "imgs")
}

func (m *Manager) baseAnnotationPath(ann *model.Annotation) string {
	return path.Join(m.baseDir, ann.DatasetID, "annotations", ann.ID)
}

func (m *Manager) DatasetAnnotationAltoDir(ann *model.Annotation) string {
	return path.Join(m.baseAnnotationPath(ann), "alto")
}

func (m *Manager) DatasetAnnotationYoloDir(ann *model.Annotation) string {
	return path.Join(m.baseAnnotationPath(ann), "yolo")
}

func (m *Manager) ModelPath(model *model.Model) string {
	return path.Join(m.modelsDir, model.LocalPath)
}

func (m *Manager) TrainingDir(t *model.Training) string {
	return path.Join(m.trainingDir, t.ID)
}
