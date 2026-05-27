package filesys

import (
	"fmt"
	"path"
	"path/filepath"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
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

func (m *Manager) DatasetPDFPath(ds *model.Dataset) string {
	return path.Join(m.DatasetDir(ds.ID), fmt.Sprintf("%s_%s.pdf", ds.EditionID, ds.FacsimileID))
}

func (m *Manager) DatasetImagesDir(ds *model.Dataset) string {
	return path.Join(m.DatasetDir(ds.ID), "imgs")
}

func (m *Manager) DatasetImagesDirByID(dsID string) string {
	return path.Join(m.DatasetDir(dsID), "imgs")
}

func (m *Manager) DatasetImageVariantsDirByID(dsID string, variant string) string {
	return path.Join(m.DatasetImagesDirByID(dsID), "_variants", variant)
}

func (m *Manager) DatasetImageVariantPathByID(dsID string, variant string, sourceFilename string) string {
	base := sourceFilename[:len(sourceFilename)-len(filepath.Ext(sourceFilename))]
	return path.Join(m.DatasetImageVariantsDirByID(dsID, variant), base+".jpg")
}

func (m *Manager) baseAnnotationPath(ann *annotation.Annotation) string {
	return path.Join(m.DatasetDir(ann.DatasetID), "annotations", ann.ID)
}

func (m *Manager) DatasetAnnotationAltoDir(ann *annotation.Annotation) string {
	return path.Join(m.baseAnnotationPath(ann), "alto")
}

func (m *Manager) DatasetAnnotationYoloDir(ann *annotation.Annotation) string {
	return path.Join(m.baseAnnotationPath(ann), "yolo")
}

func (m *Manager) ModelPath(model *model.Model) string {
	return path.Join(m.modelsDir, model.LocalPath)
}

func (m *Manager) DiagramCropsMetadataFile(editionKey string) string {
	return path.Join(m.diagramsDir, editionKey+".json")
}

func (m *Manager) EditionTxtTranscriptionDir(ed *model.Edition) string {
	return path.Join(m.baseDir, "transcriptions", ed.Key)
}

func (m *Manager) AnnotationTxtTranscriptionDir(ann *annotation.Annotation) string {
	return path.Join(m.baseAnnotationPath(ann), "transcriptions")
}

func (m *Manager) EditionTxtPageTranscriptionDir(ed *model.Edition, key string) string {
	var d = key
	if pageNum, err := strconv.Atoi(key); err == nil {
		d = pagesparser.PageToFilename(pageNum, "")
	}
	return path.Join(m.EditionTxtTranscriptionDir(ed), d)
}

func (m *Manager) AnnotationTxtPageTranscriptionDir(ann *annotation.Annotation, key string) string {
	var d = key
	if pageNum, err := strconv.Atoi(key); err == nil {
		d = pagesparser.PageToFilename(pageNum, "")
	}
	return path.Join(m.AnnotationTxtTranscriptionDir(ann), d)
}
