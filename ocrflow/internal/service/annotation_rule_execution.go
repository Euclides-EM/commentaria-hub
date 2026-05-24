package service

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/job"
)

type AnnotationRuleExecution struct {
	annotations *Annotation
	jobs        *Job
}

func NewAnnotationRuleExecution(annotations *Annotation, jobs *Job) *AnnotationRuleExecution {
	return &AnnotationRuleExecution{
		annotations: annotations,
		jobs:        jobs,
	}
}

func (e *AnnotationRuleExecution) ApplyRules(datasetID string, annotationID string, rules *annotationrule.ApplyRules) (any, error) {
	if rules.ExecutionMode != annotationrule.ExecutionModeAsync {
		return e.annotations.ApplyRules(datasetID, annotationID, rules)
	}
	return e.jobs.CreateJob(&job.Job{
		Task: job.AnnotationRuleApply,
		Annotation: &annotation.Reference{
			DatasetID: datasetID,
			ID:        annotationID,
		},
		Rules: rules,
	})
}
