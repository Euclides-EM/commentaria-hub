package service

import (
	"fmt"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	tei2 "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei"
	model2 "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
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
	if edition == nil {
		return nil, fmt.Errorf("%w: edition with key %s does not exist", ErrEditionNotFound, editionID)
	}

	tei, err := t.getTEI(edition, pageNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get TEI for edition %s: %v", editionID, err)
	}

	xml, err := tei.ToXML()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize TEI to XML for edition %s: %v", editionID, err)
	}

	return xml, nil
}

func (t *EditionTEI) getTEI(edition *model.Edition, pageNum int) (*model2.TEI, error) {
	a, _, err := t.fileSysMgt.RetrieveEditionAltoPage(edition, pageNum)
	if err == nil {
		pageKey := fmt.Sprintf("%d", pageNum)
		return tei2.BuildTEIFromALTO(pageKey, a, nil, "", model.EditionToBiblFull(edition))
	}

	markdown, err := t.fileSysMgt.RetrieveEditionMarkdownPage(edition.Key, pageNum)
	if err == nil {
		pageKey := fmt.Sprintf("%d", pageNum)
		return tei2.BuildTEIFromMarkdown(pageKey, markdown, model.EditionToBiblFull(edition))
	}

	pageKey := fmt.Sprintf("%d", pageNum)
	lines, translations, err := t.fileSysMgt.RetrieveEditionTXTPage(edition.Key, pageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve TXT page for edition %s: %v", edition.Key, err)
	}

	pageLines := tei2.Lines{
		TranscriptionLines: lines,
		Translations:       translations,
	}
	return tei2.BuildTEIFromLines(pageKey, pageLines, nil, "", model.EditionToBiblFull(edition))
}
