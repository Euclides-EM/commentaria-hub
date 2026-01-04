package annotationrule

import (
	"encoding/json"
	"fmt"
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

type Type string

const (
	TypeSegment                      Type = "segment"
	TypeSlicePages                   Type = "slice_pages"
	TypeStretch                      Type = "stretch"
	TypeAddMargin                    Type = "add_margin"
	TypeLinesDetect                  Type = "lines_detect"
	TypeRemoveCategories             Type = "remove_categories"
	TypeRemoveOverlap                Type = "remove_overlap"
	TypeReassignTextLinesByTolerance Type = "reassign_text_lines_by_tolerance"
)

type ApplyRules struct {
	Rules  []AnnotationRule `json:"rules"`
	Action ApplyRulesAction `json:"action"`
}

type AnnotationRule interface {
	GetType() Type
}

type Base struct {
	Type Type `json:"type" example:""`
}

// -- Internal helper for UnmarshalJSON --

type annotationApplyRulesRaw struct {
	Action ApplyRulesAction  `json:"action"`
	Rules  []json.RawMessage `json:"rules"`
}

func (a *ApplyRules) UnmarshalJSON(data []byte) error {
	var raw annotationApplyRulesRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	a.Action = raw.Action
	a.Rules = make([]AnnotationRule, 0, len(raw.Rules))

	for _, r := range raw.Rules {
		// Peek at the "type" field
		var base Base
		if err := json.Unmarshal(r, &base); err != nil {
			return fmt.Errorf("read rule type: %w", err)
		}

		var rule AnnotationRule
		switch base.Type {
		case TypeSlicePages:
			var v SlicePages
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal slice_pages rule: %w", err)
			}
			rule = &v

		case TypeStretch:
			var v Stretch
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal stretch rule: %w", err)
			}
			rule = &v

		case TypeAddMargin:
			var v AddMargin
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal add_margin rule: %w", err)
			}
			rule = &v

		case TypeLinesDetect:
			var v LinesDetect
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal lines_detect rule: %w", err)
			}
			rule = &v
		case TypeSegment:
			var v Segment
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal segment rule: %w", err)
			}
			rule = &v
		case TypeRemoveCategories:
			var v RemoveCategories
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal remove_categories rule: %w", err)
			}
			rule = &v
		case TypeRemoveOverlap:
			var v RemoveOverlap
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal remove_overlap rule: %w", err)
			}
			rule = &v
		case TypeReassignTextLinesByTolerance:
			var v ReassignTextLinesByTolerance
			if err := json.Unmarshal(r, &v); err != nil {
				return fmt.Errorf("unmarshal reassign_text_lines_by_tolerance rule: %w", err)
			}
			rule = &v
		default:
			return fmt.Errorf("unknown annotation rule type %q", base.Type)
		}

		a.Rules = append(a.Rules, rule)
	}

	return nil
}
