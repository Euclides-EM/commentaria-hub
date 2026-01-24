package service

import (
	"slices"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
)

type MetadataDetails struct {
}

func NewMetadataDetails() *MetadataDetails {
	return &MetadataDetails{}
}

func (m *MetadataDetails) ListAnnotationRules() ([]*annotationrule.MetadataDetails, error) {
	var res []*annotationrule.MetadataDetails
	for _, rule := range annotationrule.AllAnnotationRuleTypes {
		mdef := annotationrule.ZeroFromType(rule)
		mdef.SetDefaultValues()
		md := &annotationrule.MetadataDetails{
			Type:    rule,
			Default: mdef,
		}
		res = append(res, md)
	}
	return res, nil
}

func (m *MetadataDetails) ListPipelineStages() ([]annotationrule.PipelineStage, error) {
	res := make([]annotationrule.PipelineStage, 0)
	for _, stage := range annotationrule.AllPipelineStages {
		res = append(res, stage)
	}
	slices.SortFunc(res, func(a, b annotationrule.PipelineStage) int {
		if a.After(b) {
			return 1
		}
		if b.After(a) {
			return -1
		}
		return 0
	})
	return res, nil
}
