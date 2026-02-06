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
}

type AnnotationRule interface {
	GetType() Type
	SetDefaultValues()
}

type Base struct {
	Type             Type            `json:"type" example:""`
	ApplicableStages []PipelineStage `json:"applicable_stages"`
}

// -- Internal helper for UnmarshalJSON --

type annotationApplyRulesRaw struct {
	Action ApplyRulesAction  `json:"action"`
	Rules  []json.RawMessage `json:"rules"`
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

func GetApplicableStages(t Type) []PipelineStage {
	return applicableStagesByType[t]
}

func MinEnsuredStage(t Type) PipelineStage {
	return minEnsuredStageByType[t]
}
