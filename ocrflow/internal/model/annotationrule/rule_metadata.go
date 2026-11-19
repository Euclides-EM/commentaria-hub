package annotationrule

import "slices"

func GetApplicableStages(t Type) []PipelineStage {
	return slices.Clone(applicableStagesByType[t])
}

type ruleWithBase interface {
	ruleBase() *Base
}

func HydrateMetadata(rule AnnotationRule) {
	if rule == nil {
		return
	}
	if withBase, ok := rule.(ruleWithBase); ok {
		base := withBase.ruleBase()
		base.Type = rule.GetType()
		base.ApplicableStages = rule.ApplicablePipelineStages()
	}
}
