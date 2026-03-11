package service

import (
	"log"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/samber/lo"
)

const (
	titlePageDatasetID                = "tps"
	titlePageSourceAnnotationID       = "ann_1"
	titlePageExperimentAnnotationName = "Title Page Experiment"
	titlePageCorpusName               = "tps_experiment"
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
	// verify that the tps dataset and ann_1 annotation exist.
	baseAnnotation, err := p.annotationSvc.Get(titlePageDatasetID, titlePageSourceAnnotationID)
	if err != nil {
		return err
	}

	tpsExpEditions, err := p.editionSvc.ListEditions(func(e any) bool {
		edition := e.(*model.Edition)
		return lo.Contains(edition.Corpus, titlePageCorpusName)
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

	annotations, err := p.annotationSvc.ListAnnotations(titlePageDatasetID)
	if err != nil {
		return err
	}
	tpsExperimentAnn, found := lo.Find(annotations, func(ann *annotation.Annotation) bool {
		return ann != nil && ann.Name == titlePageExperimentAnnotationName
	})
	if !found {
		log.Printf("title page experiment annotation not found, creating it...")
		tpsExperimentAnn, err = p.annotationSvc.Create(titlePageDatasetID, &annotation.Annotation{
			Meta: common.Meta{
				Name: titlePageExperimentAnnotationName,
				Description: "Annotation used for the title page experiment. " +
					"These annotations are automatically generated based on the metadata of the editions in the tps_experiment corpus during the server startup.",
			},
			Pages:              strings.Join(editionKeys, ","),
			Segmented:          baseAnnotation.Segmented,
			GroundTruth:        baseAnnotation.GroundTruth,
			Ocred:              baseAnnotation.Ocred,
			DatasetID:          titlePageDatasetID,
			OriginAnnotationID: titlePageSourceAnnotationID,
			Hidden:             false,
			LinesDetected:      baseAnnotation.LinesDetected,
			PipelineStage:      baseAnnotation.PipelineStage,
		}, false)
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
	if _, err := p.annotationSvc.Update(titlePageDatasetID, tpsExperimentAnn.ID, tpsExperimentAnn); err != nil {
		return err
	}
	log.Printf("updated title page experiment annotation with new edition keys")
	return nil
}
