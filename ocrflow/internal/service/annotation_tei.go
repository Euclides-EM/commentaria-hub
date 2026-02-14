package service

import (
	"fmt"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
)

type AnnotationTEI struct {
	annotationSvc *Annotation
	fileSysMgt    *filesys.Manager
}

func NewAnnotationTEI(annotationSvc *Annotation, fileSysMgt *filesys.Manager) *AnnotationTEI {
	return &AnnotationTEI{
		annotationSvc: annotationSvc,
		fileSysMgt:    fileSysMgt,
	}
}

func (t *AnnotationTEI) GetTEI(datasetID string, annotationID string, pageNum string) ([]byte, error) {
	ann, err := t.annotationSvc.Get(datasetID, annotationID)
	if err != nil {
		return nil, err
	}

	if !ann.Ocred {
		return nil, fmt.Errorf("no OCR data for annotation %s", ann.ID)
	}

	// convert pageNum to int
	page, err := strconv.Atoi(pageNum)
	if err != nil {
		return nil, fmt.Errorf("invalid page number %s: %w", pageNum, err)
	}

	a, _, err := t.fileSysMgt.RetrieveAltoPage(ann, page)
	if err != nil {
		return nil, err
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
