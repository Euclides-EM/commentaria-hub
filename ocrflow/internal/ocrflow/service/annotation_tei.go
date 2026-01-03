package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
)

type AnnotationTEI struct {
	annotationSvc *Annotation
	datasetSvc    *Dataset
}

func (t *AnnotationTEI) GetTEI(datasetID string, annotationID string, pageNum string) ([]byte, error) {
	ann, err := t.annotationSvc.Get(datasetID, annotationID)
	if err != nil {
		return nil, err
	}

	if ann.AltoDir == "" {
		return nil, fmt.Errorf("no ALTO annotations found for annotation %s", ann.ID)
	}

	// convert pageNum to int
	page, err := strconv.Atoi(pageNum)
	if err != nil {
		return nil, fmt.Errorf("invalid page number %s: %w", pageNum, err)
	}

	pageAltoPath := filepath.Join(ann.AltoDir, pagesparser.PageToXMLFilename(page))
	if _, err := os.Stat(pageAltoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("page ALTO %s does not exist for annotation %s", pageAltoPath, ann.ID)
	}

	a, err := alto.LoadFromFile(pageAltoPath)
	if err != nil {
		return nil, fmt.Errorf("load ALTO: %w", err)
	}

	opts := formatcov.ALTOToTEIOptions{
		RowTolPx:     6,
		ParaGapPx:    28,
		KeepEmpty:    false,
		Title:        "Converted from ALTO",
		FacsFromPage: true,
	}

	teiData, err := formatcov.ConvertALTOToTEI(a, opts)
	if err != nil {
		return nil, err
	}

	return teiData, nil
}

func NewAnnotationTEI(annotationSvc *Annotation, datasetSvc *Dataset) *AnnotationTEI {
	return &AnnotationTEI{
		annotationSvc: annotationSvc,
		datasetSvc:    datasetSvc,
	}
}
