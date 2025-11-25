package store

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"path"
)

// Dataset storage layout:
//
// <baseDir>
// └─ <dataset_id>
//    ├─ <edition_id>_<facsimile_id>.pdf
//    ├─ imgs
//    │  ├─ 0001.jpg
//    │  ├─ 0002.jpg
//    │  └─ ...
//    └─ annotations
//       └─ <annotation_id>

func DatasetPDFPath(ds *model.Dataset, baseDir string) string {
	return path.Join(baseDir, ds.ID, fmt.Sprintf("%s_%s.pdf", ds.EditionID(), ds.FacsimileID()))
}

func DatasetImagesDir(ds *model.Dataset, baseDir string) string {
	return path.Join(baseDir, ds.ID, "imgs")
}

func DatasetAnnotationsPath(ann *model.Annotation, baseDir string) string {
	return path.Join(baseDir, ann.DatasetID(), "annotations", ann.ID)
}
