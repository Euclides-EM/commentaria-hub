package service

import (
	"fmt"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotationrule"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/job"
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
	if rules == nil {
		return nil, fmt.Errorf("missing annotation rules")
	}
	if rules.ExecutionMode != annotationrule.ExecutionModeAsync && rules.AnyRuleRequireGPUFarm() {
		return nil, fmt.Errorf("GPU farm detection rules must be run with async execution mode")
	}
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
			DatasetID: datasetID,
			ID:        annotationID,
		},
		Target: &job.Target{
			DatasetID:    datasetID,
			AnnotationID: ann.ID,
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
