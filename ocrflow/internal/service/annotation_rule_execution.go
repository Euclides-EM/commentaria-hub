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
	ann, err := e.annotations.PrepareApplyRules(datasetID, annotationID, rules)
	if err != nil {
		return nil, err
	}

	if rules.ExecutionMode != annotationrule.ExecutionModeAsync {
		return e.annotations.ExecuteApplyRules(ann.DatasetID, ann.ID, rules)
	}

	jobResult, err := e.jobs.CreateJob(&job.Job{
		Task: job.AnnotationRuleApply,
		Annotation: &annotation.Reference{
			DatasetID: ann.DatasetID,
			ID:        ann.ID,
		},
		Rules: rules,
	})
	if err != nil {
		return nil, err
	}
	if rules.Action == annotationrule.ApplyRulesActionCreateNew {
		return ann, nil
	}
	return jobResult, nil
}
