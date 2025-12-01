package store

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"path"
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

func DatasetPDFPath(ds *model.Dataset, baseDir string) string {
	return path.Join(baseDir, ds.ID, fmt.Sprintf("%s_%s.pdf", ds.EditionID(), ds.FacsimileID()))
}

func DatasetImagesDir(ds *model.Dataset, baseDir string) string {
	return path.Join(baseDir, ds.ID, "imgs")
}

func datasetAnnotationsPath(ann *model.Annotation, baseDir string) string {
	return path.Join(baseDir, ann.DatasetID(), "annotations", ann.ID)
}

func DatasetAnnotationAltoDir(ann *model.Annotation, baseDir string) string {
	return path.Join(datasetAnnotationsPath(ann, baseDir), "alto")
}

func DatasetAnnotationYoloDir(ann *model.Annotation, baseDir string) string {
	return path.Join(datasetAnnotationsPath(ann, baseDir), "yolo")
}

func DatasetAnnotationRoboflowDir(ann *model.Annotation, baseDir string) string {
	return path.Join(datasetAnnotationsPath(ann, baseDir), "roboflow")
}

func TrainingDir(t *model.Training, dir string) string {
	return path.Join(dir, t.ID)
}
