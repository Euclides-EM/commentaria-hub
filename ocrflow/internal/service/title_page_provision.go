package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/titlepage"
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
	if err := p.upsertAnnotationForCorpus(titlepage.ExperimentCorpusName, titlepage.ExperimentAnnotationName); err != nil {
		return err
	}
	if err := p.upsertAnnotationForCorpus(titlepage.ReviewedCorpusName, titlepage.ReviewedAnnotationName); err != nil {
		return err
	}
	return nil
}

func (p *TitlePageProvision) upsertAnnotationForCorpus(corpus, annotationName string) error {
	// verify that the basic Title Page dataset and annotation exist.
	baseAnnotation, err := p.annotationSvc.Get(titlepage.DatasetID, titlepage.AnnotationID)
	if err != nil {
		return err
	}

	editions, err := p.editionSvc.ListEditions(func(e any) bool {
		edition := e.(*model.Edition)
		return lo.Contains(edition.Corpus, corpus)
	}, nil, 0, 1000)
	if err != nil {
		return err
	}
	if editions.Total == 0 {
		log.Printf("warning: no editions found in the %s corpus, skipping title page annotation provisioning", corpus)
		return nil
	}
	editionKeys := lo.Map(editions.Items, func(e *model.Edition, _ int) string {
		return e.Key
	})

	annotations, err := p.annotationSvc.ListAnnotations(titlepage.DatasetID)
	if err != nil {
		return err
	}
	ann, found := lo.Find(annotations, func(ann *annotation.Annotation) bool {
		return ann != nil && ann.Name == annotationName
	})
	if !found {
		log.Printf("title page experiment annotation not found, creating it...")
		ann, err = p.annotationSvc.Create(titlepage.DatasetID, &annotation.Annotation{
			Meta: common.Meta{
				Name: annotationName,
				Description: fmt.Sprintf("Annotation used for the %s. "+
					"These annotations are automatically generated based on the metadata of the editions in the %s corpus during the server startup.", strings.ToLower(annotationName), corpus),
			},
			Pages:              strings.Join(editionKeys, ","),
			Segmented:          baseAnnotation.Segmented,
			GroundTruth:        baseAnnotation.GroundTruth,
			Ocred:              baseAnnotation.Ocred,
			DatasetID:          titlepage.DatasetID,
			OriginAnnotationID: titlepage.AnnotationID,
			Hidden:             false,
			LinesDetected:      baseAnnotation.LinesDetected,
			PipelineStage:      baseAnnotation.PipelineStage,
		}, false)
		if err != nil {
			return err
		}
		log.Printf("created annotation %s with ID %s", annotationName, ann.ID)
	}
	editionsInAnn := strings.Split(ann.Pages, ",")
	if lo.ElementsMatch(editionKeys, editionsInAnn) {
		log.Printf("%s annotation is up to date, no update needed", annotationName)
	}
	log.Printf("updating %s annotation with new edition keys...", annotationName)
	ann.Pages = strings.Join(editionKeys, ",")
	if _, err := p.annotationSvc.Update(titlepage.DatasetID, ann.ID, ann); err != nil {
		return err
	}
	log.Printf("updated %s annotation with new edition keys", annotationName)
	return nil
}
