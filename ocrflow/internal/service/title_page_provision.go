package service

import (
	"log"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/samber/lo"
)

const (
	titlePageDatasetID          = "tps"
	titlePageSourceAnnotationID = "ann_1"
	titlePageExperimentAnnID    = "ann_experiment"
)

type TitlePageProvision struct {
	annotationSvc    *Annotation
	datasetSvc       *Dataset
	editionSvc       *Edition
	featureResultSvc *Result
}

func NewTitlePageProvision(annotationSvc *Annotation, datasetSvc *Dataset, editionSvc *Edition, featureResultSvc *Result) *TitlePageProvision {
	return &TitlePageProvision{
		annotationSvc:    annotationSvc,
		datasetSvc:       datasetSvc,
		editionSvc:       editionSvc,
		featureResultSvc: featureResultSvc,
	}
}

func (p *TitlePageProvision) UpdateTitlePageAnnotationsByMetadataInfo() error {
	// verify that the tps dataset and ann_1 annotation exist.
	_, err := p.annotationSvc.Get(titlePageDatasetID, titlePageSourceAnnotationID)
	if err != nil {
		return err
	}

	tpsExpEditions, err := p.editionSvc.ListEditions(func(e any) bool {
		edition := e.(*model.Edition)
		return lo.Contains(edition.Corpus, "tps_experiment")
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

	tpsExperimentAnn, err := p.annotationSvc.Get(titlePageDatasetID, titlePageExperimentAnnID)
	if err != nil {
		log.Printf("title page experiment annotation not found, creating it...")
		tpsExperimentAnn, err = p.annotationSvc.Create(titlePageDatasetID, &annotation.Annotation{
			Meta: common.Meta{
				ID:   titlePageExperimentAnnID,
				Name: "Title page experiment",
				Description: "Annotation used for the title page experiment. " +
					"These annotations are automatically generated based on the metadata of the editions in the tps_experiment corpus during the server startup.",
			},
			Pages:              strings.Join(editionKeys, ","),
			Segmented:          false,
			GroundTruth:        false,
			Ocred:              true,
			DatasetID:          titlePageDatasetID,
			OriginAnnotationID: titlePageSourceAnnotationID,
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
	if _, err := p.annotationSvc.Update(titlePageDatasetID, tpsExperimentAnn.ID, tpsExperimentAnn); err != nil {
		return err
	}
	log.Printf("updated title page experiment annotation with new edition keys")
	log.Printf("copying feature results from ann_1 to %s for %d edition keys...", tpsExperimentAnn.ID, len(editionKeys))
	if err := p.copyFeatureResults(editionKeys, tpsExperimentAnn.ID); err != nil {
		return err
	}
	log.Printf("copied feature results from ann_1 to %s for %d keys", tpsExperimentAnn.ID, len(editionKeys))
	return nil
}

func (p *TitlePageProvision) copyFeatureResults(keys []string, targetAnnotationID string) error {
	if len(keys) == 0 {
		return nil
	}
	if targetAnnotationID == "" {
		return nil
	}

	sourceResults, err := p.featureResultSvc.ListResults(titlePageDatasetID, titlePageSourceAnnotationID, keys, nil)
	if err != nil {
		return err
	}
	if len(sourceResults) == 0 {
		return nil
	}

	existingResults, err := p.featureResultSvc.ListResults(titlePageDatasetID, targetAnnotationID, keys, nil)
	if err != nil {
		return err
	}
	existingByFeatureAndKey := lo.SliceToMap(existingResults, func(r *feature.Result) (string, struct{}) {
		return r.FeatureID + "|" + r.PageKey, struct{}{}
	})

	cloned := make([]*feature.Result, 0, len(sourceResults))
	for _, result := range sourceResults {
		if result == nil {
			continue
		}
		if _, exists := existingByFeatureAndKey[result.FeatureID+"|"+result.PageKey]; exists {
			continue
		}

		copied := *result
		copied.AnnotationID = targetAnnotationID
		copied.Values = make([]feature.ResultValue, 0, len(result.Values))
		for _, value := range result.Values {
			copiedValue := feature.ResultValue{
				Surface: value.Surface,
			}
			if len(value.Properties) > 0 {
				copiedValue.Properties = mapsClone(value.Properties)
			}
			copied.Values = append(copied.Values, copiedValue)
		}
		cloned = append(cloned, &copied)
	}

	if len(cloned) == 0 {
		return nil
	}

	return p.featureResultSvc.CreateResults(cloned)
}

func mapsClone(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	return lo.Assign(map[string]string{}, m)
}
