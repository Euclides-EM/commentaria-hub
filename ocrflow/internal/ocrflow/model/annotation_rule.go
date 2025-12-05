package model

import (
	"encoding/json"
	"fmt"
)

type AnnotationApplyRulesAction string

const (
	AnnotationApplyRulesActionCreateNew AnnotationApplyRulesAction = "create_new"
	AnnotationApplyRulesActionOverwrite AnnotationApplyRulesAction = "overwrite"
)

type AnnotationRuleContactType string

const (
	AnnotationRuleContactTypeInner AnnotationRuleContactType = "inner"
	AnnotationRuleContactTypeOuter AnnotationRuleContactType = "outer"
)

type AnnotationRuleContactSide string

const (
	AnnotationRuleContactSideLeft       AnnotationRuleContactSide = "left"
	AnnotationRuleContactSideRight      AnnotationRuleContactSide = "right"
	AnnotationRuleContactSideTop        AnnotationRuleContactSide = "top"
	AnnotationRuleContactSideBottom     AnnotationRuleContactSide = "bottom"
	AnnotationRuleContactSideHorizontal AnnotationRuleContactSide = "horizontal"
	AnnotationRuleContactSideVertical   AnnotationRuleContactSide = "vertical"
	AnnotationRuleContactSideAll        AnnotationRuleContactSide = "all"
)

type AnnotationRuleType string

const (
	AnnotationRuleTypeSlicePages  AnnotationRuleType = "slice_pages"
	AnnotationRuleTypeStretch     AnnotationRuleType = "stretch"
	AnnotationRuleTypeAddMargin   AnnotationRuleType = "add_margin"
	AnnotationRuleTypeLinesDetect AnnotationRuleType = "lines_detect"
)

type AnnotationApplyRules struct {
	Rules  []AnnotationRule           `json:"rules"`
	Action AnnotationApplyRulesAction `json:"action"`
}

type AnnotationRule interface {
	GetType() AnnotationRuleType
	DeepCopy() AnnotationRule
}

type AnnotationRuleBase struct {
	Type AnnotationRuleType `json:"type"`
}

type AnnotationRuleSlicePages struct {
	AnnotationRuleBase `json:",inline"`
	Pages              string `json:"pages"`
}

func (t *AnnotationRuleSlicePages) GetType() AnnotationRuleType {
	return AnnotationRuleTypeSlicePages
}

func (t *AnnotationRuleSlicePages) DeepCopy() AnnotationRule {
	if t == nil {
		return nil
	}
	return &AnnotationRuleSlicePages{
		AnnotationRuleBase: AnnotationRuleBase{Type: t.Type},
		Pages:              t.Pages,
	}
}

type AnnotationRuleLinesDetect struct {
	AnnotationRuleBase `json:",inline"`
	// IncludeCategories specifies which categories to run line detection on. For example, "MainZone".
	IncludeCategories []string `json:"include_categories,omitempty"`
	// IgnoreCategories specifies which categories to ignore when running line detection. For example, "GraphicZone", "DigitizationArtefactZone", ...
	IgnoreCategories []string `json:"ignore_categories,omitempty"`
}

func (t *AnnotationRuleLinesDetect) GetType() AnnotationRuleType {
	return AnnotationRuleTypeLinesDetect
}

func (t *AnnotationRuleLinesDetect) DeepCopy() AnnotationRule {
	if t == nil {
		return nil
	}
	return &AnnotationRuleLinesDetect{
		AnnotationRuleBase: AnnotationRuleBase{Type: t.Type},
		IncludeCategories:  append([]string{}, t.IncludeCategories...),
		IgnoreCategories:   append([]string{}, t.IgnoreCategories...),
	}
}

type AnnotationRuleStretch struct {
	AnnotationRuleBase `json:",inline"`
	StretchCategory    string                    `json:"stretch_category"`
	Towards            string                    `json:"towards"`
	ContactType        AnnotationRuleContactType `json:"contact_type"`
	ContactSide        AnnotationRuleContactSide `json:"contact_side"`
}

func (t *AnnotationRuleStretch) GetType() AnnotationRuleType {
	return AnnotationRuleTypeStretch
}

func (t *AnnotationRuleStretch) DeepCopy() AnnotationRule {
	if t == nil {
		return nil
	}
	return &AnnotationRuleStretch{
		AnnotationRuleBase: AnnotationRuleBase{Type: t.Type},
		StretchCategory:    t.StretchCategory,
		Towards:            t.Towards,
		ContactType:        t.ContactType,
		ContactSide:        t.ContactSide,
	}
}

type AnnotationRuleAddMargin struct {
	AnnotationRuleBase `json:",inline"`
	Margin             float64                   `json:"margin"`
	Side               AnnotationRuleContactSide `json:"sides"`
	Category           string                    `json:"category"`
}

func (t *AnnotationRuleAddMargin) GetType() AnnotationRuleType {
	return AnnotationRuleTypeAddMargin
}

func (t *AnnotationRuleAddMargin) DeepCopy() AnnotationRule {
	if t == nil {
		return nil
	}
	return &AnnotationRuleAddMargin{
		AnnotationRuleBase: AnnotationRuleBase{Type: t.Type},
		Margin:             t.Margin,
		Side:               t.Side,
		Category:           t.Category,
	}
}

func NewAnnotationRuleSlicePages(pages string) *AnnotationRuleSlicePages {
	return &AnnotationRuleSlicePages{
		AnnotationRuleBase: AnnotationRuleBase{Type: AnnotationRuleTypeSlicePages},
		Pages:              pages,
	}
}

func NewAnnotationRuleStretch(cat, towards string, ct AnnotationRuleContactType, side AnnotationRuleContactSide) *AnnotationRuleStretch {
	return &AnnotationRuleStretch{
		AnnotationRuleBase: AnnotationRuleBase{Type: AnnotationRuleTypeStretch},
		StretchCategory:    cat,
		Towards:            towards,
		ContactType:        ct,
		ContactSide:        side,
	}
}

func NewAnnotationRuleAddMargin(margin float64, side AnnotationRuleContactSide) *AnnotationRuleAddMargin {
	return &AnnotationRuleAddMargin{
		AnnotationRuleBase: AnnotationRuleBase{Type: AnnotationRuleTypeAddMargin},
		Margin:             margin,
		Side:               side,
	}
}

func NewAnnotationRuleLinesDetect(detect bool) *AnnotationRuleLinesDetect {
	return &AnnotationRuleLinesDetect{
		AnnotationRuleBase: AnnotationRuleBase{Type: AnnotationRuleTypeLinesDetect},
	}
}

// Internal helper for UnmarshalJSON
type annotationApplyRulesRaw struct {
	Action AnnotationApplyRulesAction `json:"action"`
	Rules  []json.RawMessage          `json:"rules"`
}

func (a *AnnotationApplyRules) UnmarshalJSON(data []byte) error {
	var raw annotationApplyRulesRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	a.Action = raw.Action
	a.Rules = make([]AnnotationRule, 0, len(raw.Rules))

	for _, r := range raw.Rules {
		// Peek at the "type" field
		var base AnnotationRuleBase
		if err := json.Unmarshal(r, &base); err != nil {
			return fmt.Errorf("read rule type: %w", err)
		}

		var rule AnnotationRule
		switch base.Type {
		case AnnotationRuleTypeSlicePages:
			var v AnnotationRuleSlicePages
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal slice_pages rule: %w", err)
			}
			rule = &v

		case AnnotationRuleTypeStretch:
			var v AnnotationRuleStretch
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal stretch rule: %w", err)
			}
			rule = &v

		case AnnotationRuleTypeAddMargin:
			var v AnnotationRuleAddMargin
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal add_margin rule: %w", err)
			}
			rule = &v

		case AnnotationRuleTypeLinesDetect:
			var v AnnotationRuleLinesDetect
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal lines_detect rule: %w", err)
			}
			rule = &v

		default:
			return fmt.Errorf("unknown annotation rule type %q", base.Type)
		}

		a.Rules = append(a.Rules, rule)
	}

	return nil
}

func (a *AnnotationApplyRules) DeepCopy() *AnnotationApplyRules {
	if a == nil {
		return nil
	}
	copied := &AnnotationApplyRules{
		Action: a.Action,
		Rules:  make([]AnnotationRule, len(a.Rules)),
	}
	for i, r := range a.Rules {
		copied.Rules[i] = r.DeepCopy()
	}
	return copied
}
