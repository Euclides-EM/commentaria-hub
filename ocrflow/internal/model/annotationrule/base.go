package annotationrule

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ApplyRulesAction string

const (
	ApplyRulesActionCreateNew ApplyRulesAction = "create_new"
	ApplyRulesActionOverwrite ApplyRulesAction = "overwrite"
)

type ContactType string

const (
	ContactTypeInner ContactType = "inner"
	ContactTypeOuter ContactType = "outer"
)

type ContactSide string

const (
	ContactSideLeft       ContactSide = "left"
	ContactSideRight      ContactSide = "right"
	ContactSideTop        ContactSide = "top"
	ContactSideBottom     ContactSide = "bottom"
	ContactSideHorizontal ContactSide = "horizontal"
	ContactSideVertical   ContactSide = "vertical"
	ContactSideAll        ContactSide = "all"
)

type ApplyRules struct {
	Rules  AnnotationRules  `json:"rules"`
	Action ApplyRulesAction `json:"action"`
	// Name is used only if the action is ApplyRulesActionCreateNew
	Name string `json:"name"`
	// Description is used only if the action is ApplyRulesActionCreateNew
	Description string `json:"description"`
	// CopyFeatureResults is used only if the action is ApplyRulesActionCreateNew. If true, the feature results of the original annotation will be copied to the new annotation.
	CopyFeatureResults bool          `json:"copy_feature_results"`
	ExecutionMode      ExecutionMode `json:"execution_mode,omitempty"`
}

func (r *ApplyRules) AnyRuleRequireGPUFarm() bool {
	for _, rule := range r.Rules {
		switch t := rule.(type) {
		case *LinesDetect:
			if t.UseGPUFarm {
				return true
			}
		case *ModelDetect:
			if t.UseGPUFarm {
				return true
			}
		}
	}
	return false
}

type ExecutionMode string

const (
	ExecutionModeSync  ExecutionMode = "sync"
	ExecutionModeAsync ExecutionMode = "async"
)

func ToExecutionMode(value string, defaultVal ExecutionMode) ExecutionMode {
	switch ExecutionMode(strings.ToLower(strings.TrimSpace(value))) {
	case ExecutionModeAsync:
		return ExecutionModeAsync
	case ExecutionModeSync:
		return ExecutionModeSync
	default:
		return defaultVal
	}
}

type AnnotationRule interface {
	GetType() Type
	SetDefaultValues()
	ApplicablePipelineStages() []PipelineStage
	EnsuredPipelineStage() PipelineStage
}

type Base struct {
	Type             Type            `json:"type" example:""`
	ApplicableStages []PipelineStage `json:"applicable_stages"`
}

func (b *Base) ruleBase() *Base {
	return b
}

func (b *Base) ApplicablePipelineStages() []PipelineStage {
	return applicableStagesByType[b.Type]
}

func (b *Base) EnsuredPipelineStage() PipelineStage {
	return minEnsuredStageByType[b.Type]
}

func UnmarshalRuleJSON(data []byte) (AnnotationRule, error) {
	var base Base
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("read rule type: %w", err)
	}

	newRule, ok := ruleFactories[base.Type]
	if !ok {
		return nil, fmt.Errorf("unknown annotation rule type %q", base.Type)
	}

	rule := newRule()
	if err := json.Unmarshal(data, rule); err != nil {
		return nil, fmt.Errorf("unmarshal %s rule: %w", base.Type, err)
	}
	HydrateMetadata(rule)

	return rule, nil
}

func ToApplyRulesAction(value string, defaultVal ApplyRulesAction) ApplyRulesAction {
	switch ApplyRulesAction(strings.ToLower(strings.TrimSpace(value))) {
	case ApplyRulesActionCreateNew:
		return ApplyRulesActionCreateNew
	case ApplyRulesActionOverwrite:
		return ApplyRulesActionOverwrite
	default:
		return defaultVal
	}
}
