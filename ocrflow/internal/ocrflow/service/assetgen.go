package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
)

type AssetGen struct {
	dataset       *Dataset
	annotationTEI *AnnotationTEI
	annotation    *Annotation
	fileSysMgt    *filesys.Manager
}

func NewAssetGen(dataset *Dataset, annotationTEI *AnnotationTEI, annotation *Annotation, fileSysMgt *filesys.Manager) *AssetGen {
	return &AssetGen{
		dataset:       dataset,
		annotationTEI: annotationTEI,
		annotation:    annotation,
		fileSysMgt:    fileSysMgt,
	}
}

func (ag *AssetGen) GenerateAssets(datasetId, annotationId string, pages []int) (zipPath string, err error) {
	ds, err := ag.dataset.Get(datasetId)
	if err != nil {
		return "", err
	}

	ann, err := ag.annotation.Get(datasetId, annotationId)
	if err != nil {
		return "", err
	}

	outDir, err := os.MkdirTemp("", "annotation_assets_*")
	if err != nil {
		return "", err
	}

	if pages == nil {
		pages, err = pagesparser.Parse(ann.Pages)
		if err != nil {
			return "", err
		}
	}

	if err := ag.generateAssetsForPages(ds, datasetId, annotationId, pages, outDir); err != nil {
		return "", err
	}
	defer os.RemoveAll(outDir)

	zipPath = path.Join(os.TempDir(), fmt.Sprintf("annotation_%s_assets_%d.zip", annotationId, time.Now().UnixNano()))
	err = futils.Zip(outDir, zipPath)
	if err != nil {
		return "", err
	}

	return zipPath, nil
}

func (ag *AssetGen) generateAssetsForPages(ds *model.Dataset, datasetId, annotationId string, pages []int, outDir string) error {
	if err := os.MkdirAll(path.Join(outDir, "imgs"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(path.Join(outDir, "tei"), 0755); err != nil {
		return err
	}

	for _, pageNum := range pages {
		tei, err := ag.annotationTEI.GetTEI(datasetId, annotationId, fmt.Sprintf("%d", pageNum))
		if err != nil {
			return err
		}
		teiOutPath := path.Join(outDir, "tei", fmt.Sprintf("page_%d.xml", pageNum))
		if err := os.WriteFile(teiOutPath, tei, 0644); err != nil {
			return err
		}
		if err := futils.CopyFile(path.Join(ag.fileSysMgt.DatasetImagesDir(ds), pagesparser.PageToPNGFilename(pageNum)), path.Join(outDir, "imgs", pagesparser.PageToPNGFilename(pageNum))); err != nil {
			return err
		}
	}

	index, err := ag.annotation.GetAnnotationIndex(datasetId, annotationId, nil)
	if err != nil {
		return err
	}

	w, err := os.Create(path.Join(outDir, "annotation_index.json"))
	if err != nil {
		return err
	}
	defer w.Close()
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(index); err != nil {
		return err
	}

	now := time.Now()
	metadata := struct {
		DatasetID    string `json:"dataset_id"`
		AnnotationID string `json:"annotation_id"`
		TotalPages   int    `json:"total_pages"`
		Date         string `json:"date"`
	}{
		DatasetID:    datasetId,
		AnnotationID: annotationId,
		TotalPages:   len(pages),
		Date:         now.Format(time.RFC3339),
	}

	metadataOutPath := path.Join(outDir, "metadata.json")
	metadataFile, err := os.Create(metadataOutPath)
	if err != nil {
		return err
	}
	defer metadataFile.Close()
	metadataEnc := json.NewEncoder(metadataFile)
	metadataEnc.SetIndent("", "  ")
	if err := metadataEnc.Encode(metadata); err != nil {
		return err
	}

	return nil

}
