package filesys

import (
	"fmt"
	"path"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
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

func (m *Manager) DatasetDir(dsID string) string {
	return path.Join(m.baseDir, dsID)
}

func (m *Manager) DatasetPDFPath(ds *ocrflow.Dataset) string {
	return path.Join(m.DatasetDir(ds.ID), fmt.Sprintf("%s_%s.pdf", ds.EditionID, ds.FacsimileID))
}

func (m *Manager) DatasetImagesDir(ds *ocrflow.Dataset) string {
	return path.Join(m.DatasetDir(ds.ID), "imgs")
}

func (m *Manager) baseAnnotationPath(ann *ocrflow.Annotation) string {
	return path.Join(m.DatasetDir(ann.DatasetID), "annotations", ann.ID)
}

func (m *Manager) DatasetAnnotationAltoDir(ann *ocrflow.Annotation) string {
	return path.Join(m.baseAnnotationPath(ann), "alto")
}

func (m *Manager) DatasetAnnotationYoloDir(ann *ocrflow.Annotation) string {
	return path.Join(m.baseAnnotationPath(ann), "yolo")
}

func (m *Manager) ModelPath(model *ocrflow.Model) string {
	return path.Join(m.modelsDir, model.LocalPath)
}

func (m *Manager) TrainingDir(t *ocrflow.Training) string {
	return path.Join(m.trainingDir, t.ID)
}
