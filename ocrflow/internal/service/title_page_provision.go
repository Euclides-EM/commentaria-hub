package service

import (
	"log"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/samber/lo"
)

type TitlePageProvision struct {
	annotationSvc *Annotation
	datasetSvc    *Dataset
	editionSvc    *Edition
}

func NewTitlePageProvision(annotationSvc *Annotation, datasetSvc *Dataset, editionSvc *Edition) *TitlePageProvision {
	return &TitlePageProvision{
		annotationSvc: annotationSvc,
		datasetSvc:    datasetSvc,
		editionSvc:    editionSvc,
	}
}

func (p *TitlePageProvision) UpdateTitlePageAnnotationsByMetadataInfo() error {
	// verify that the tps dataset and anno_1 annotation exist.
	_, err := p.annotationSvc.Get("tps", "ann_1")
	if err != nil {
		return err
	}

	tpsExpEditions, err := p.editionSvc.ListEditions(func(e any) bool {
		edition := e.(*model.Edition)
		return lo.Contains(edition.Corpus, "tps_experiment_changeme")
	}, nil, 0, 1000)
	if err != nil {
		return err
	}
	if tpsExpEditions.Total == 0 {
		log.Printf("warning: no editions found in the tps_experiment corpus, skipping title page annotation provisioning")
		return nil
	}
	editionKeys := lo.Map(tpsExpEditions.Items, func(e *model.Edition, _ int) string {
		return e.Key
	})

	tpsExperimentAnn, err := p.annotationSvc.Get("tps", "ann_experiment")
	if err != nil {
		log.Printf("title page experiment annotation not found, creating it...")
		tpsExperimentAnn, err = p.annotationSvc.Create("tps", &annotation.Annotation{
			Meta: common.Meta{
				ID:   "ann_experiment",
				Name: "Title page experiment",
				Description: "Annotation used for the title page experiment. " +
					"These annotations are automatically generated based on the metadata of the editions in the tps_experiment corpus during the server startup.",
			},
			Pages:              strings.Join(editionKeys, ","),
			Segmented:          false,
			GroundTruth:        false,
			Ocred:              true,
			DatasetID:          "tps",
			OriginAnnotationID: "ann_1",
		})
		if err != nil {
			return err
		}
		log.Printf("created title page experiment annotation with ID %s", tpsExperimentAnn.ID)
	}
	editionsInAnn := strings.Split(tpsExperimentAnn.Pages, ",")
	if lo.ElementsMatch(editionKeys, editionsInAnn) {
		log.Printf("title page experiment annotation is up to date, no update needed")
	}
	log.Printf("updating title page experiment annotation with new edition keys...")
	tpsExperimentAnn.Pages = strings.Join(editionKeys, ",")
	if _, err := p.annotationSvc.Update("tps", "ann_experiment", tpsExperimentAnn); err != nil {
		return err
	}
	log.Printf("updated title page experiment annotation with new edition keys")
	return nil
}
