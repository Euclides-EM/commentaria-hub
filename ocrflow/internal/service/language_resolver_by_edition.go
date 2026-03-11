package service

import (
	"fmt"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/titlepage"
	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type LanguagesResolver struct {
	editionSvc *Edition
	datasetSvc *Dataset
}

func NewLanguagesResolver(editionSvc *Edition, datasetSvc *Dataset) *LanguagesResolver {
	return &LanguagesResolver{
		editionSvc: editionSvc,
		datasetSvc: datasetSvc,
	}
}

func (lr *LanguagesResolver) Resolve(datasetID string, key string) ([]string, error) {
	var editionID string
	if datasetID != titlepage.DatasetID {
		ds, err := lr.datasetSvc.Get(datasetID)
		if err != nil {
			return nil, err
		}
		editionID = ds.EditionID
	} else {
		editionID = key
	}

	if editionID == "" {
		return nil, fmt.Errorf("could not link dataset %s and key %s to an edition, currently only edition-linked datasets are supported for language resolution", datasetID, key)
	}

	edition, err := lr.editionSvc.GetEditionByID(editionID)
	if err != nil {
		return nil, err
	}
	title := cases.Title(language.Und)
	languages := lo.FilterMap(edition.Languages, func(lang string, _ int) (string, bool) {
		lang = strings.TrimSpace(lang)
		if lang == "" {
			return "", false
		}
		return title.String(lang), true
	})
	return languages, nil
}
