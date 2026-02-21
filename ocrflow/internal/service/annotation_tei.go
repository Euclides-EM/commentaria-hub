package service

import (
	"fmt"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
)

type AnnotationTEI struct {
	annotationSvc *Annotation
	fileSysMgt    *filesys.Manager
	titlePageTEI  *TitlePageTEI
}

func NewAnnotationTEI(annotationSvc *Annotation, fileSysMgt *filesys.Manager, titlePageTEI *TitlePageTEI) *AnnotationTEI {
	return &AnnotationTEI{
		annotationSvc: annotationSvc,
		fileSysMgt:    fileSysMgt,
		titlePageTEI:  titlePageTEI,
	}
}

func (t *AnnotationTEI) GetTEI(datasetID string, annotationID string, pageNumOrKey string, features []string) ([]byte, error) {
	ann, err := t.annotationSvc.Get(datasetID, annotationID)
	if err != nil {
		return nil, err
	}

	if ann.Ocred {
		return t.getTEIFromALTO(ann, pageNumOrKey)
	}

	if datasetID == "tps" && annotationID == "ann_1" {
		return t.titlePageTEI.GetTEI(datasetID, annotationID, pageNumOrKey, features)
	}

	return nil, fmt.Errorf("no OCR data for annotation %s", ann.ID)
}

func (t *AnnotationTEI) getTEIFromALTO(ann *annotation.Annotation, pageNumOrKey string) ([]byte, error) {
	// convert pageNumOrKey to int
	page, err := strconv.Atoi(pageNumOrKey)
	if err != nil {
		return nil, fmt.Errorf("invalid page number %s: %w", pageNumOrKey, err)
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
