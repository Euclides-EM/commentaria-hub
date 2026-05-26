package service

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/job"
	"github.com/tiendc/go-deepcopy"
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

	targetAnnotationID := annotationID
	jobRules := rules
	if rules.Action == annotationrule.ApplyRulesActionCreateNew {
		ann, err := e.annotations.prepareApplyRulesTarget(datasetID, annotationID, rules)
		if err != nil {
			return nil, err
		}
		targetAnnotationID = ann.ID

		jobRules = &annotationrule.ApplyRules{}
		if err := deepcopy.Copy(&jobRules, &rules); err != nil {
			return nil, err
		}
		jobRules.Action = annotationrule.ApplyRulesActionOverwrite
		jobRules.Name = ""
		jobRules.Description = ""
		jobRules.CopyFeatureResults = false
	}

	jobResult, err := e.jobs.CreateJob(&job.Job{
		Task: job.AnnotationRuleApply,
		Annotation: &annotation.Reference{
			DatasetID: datasetID,
			ID:        targetAnnotationID,
		},
		Rules: jobRules,
	})
	if err != nil {
		return nil, err
	}
	if rules.Action == annotationrule.ApplyRulesActionCreateNew {
		return e.annotations.Get(datasetID, targetAnnotationID)
	}
	return jobResult, nil
}
