package service

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
)

type EditionTEI struct {
	fileSysMgt *filesys.Manager
	editionSvc *Edition
}

func NewEditionTEI(fileSysMgt *filesys.Manager, editionSvc *Edition) *EditionTEI {
	return &EditionTEI{
		fileSysMgt: fileSysMgt,
		editionSvc: editionSvc,
	}
}

func (t *EditionTEI) GetTEI(editionID string, pageNum int) ([]byte, error) {
	edition, err := t.editionSvc.GetEditionByID(editionID)
	if err != nil {
		return nil, err
	}

	return t.getTEI(edition, pageNum)
}

func (t *EditionTEI) getTEI(edition *model.Edition, pageNum int) ([]byte, error) {
	a, _, err := t.fileSysMgt.RetrieveEditionAltoPage(edition, pageNum)
	if err == nil {
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

	originalLines, linesByLang, err := t.fileSysMgt.RetrieveEditionTXTPage(edition, pageNum)
	if err != nil {
		return nil, err
	}

	teiData, err := formatcov.ConvertTXTToTEI(originalLines, linesByLang)
	if err != nil {
		return nil, err
	}

	return teiData, nil
}
