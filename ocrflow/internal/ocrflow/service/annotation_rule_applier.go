package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/coco"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/roboflow"
	"github.com/samber/lo"
	"log"
)

type AnnotationRuleApplier struct {
}

func (a *AnnotationRuleApplier) ApplyRules(ann *model.Annotation, rules []model.AnnotationRule) error {
	var err error
	log.Printf("Applying %d rules to annotation %s", len(rules), ann.ID)
	for _, rule := range rules {
		log.Printf("Applying rule %+v", rule)
		ann, err = a.ApplyRule(ann, rule)
		if err != nil {
			return fmt.Errorf("failed to apply rule %+v: %w", rule, err)
		}
	}
	log.Printf("Applied %d rules", len(rules))
	return nil
}

func (a *AnnotationRuleApplier) ApplyRule(ann *model.Annotation, rule model.AnnotationRule) (*model.Annotation, error) {
	switch t := rule.(type) {
	case *model.AnnotationRuleSlicePages:
		return a.applySlicePagesRule(ann, t)
	case *model.AnnotationRuleStretch:
		return a.applyStretchRule(ann, t)
	case *model.AnnotationRuleAddMargin:
		return a.applyAddMarginRule(ann, t)
	default:
		log.Printf("Unknown rule type: %s", rule.GetType())
	}
	return nil, fmt.Errorf("unknown rule type: %s", rule.GetType())
}

func (a *AnnotationRuleApplier) applySlicePagesRule(ann *model.Annotation, t *model.AnnotationRuleSlicePages) (*model.Annotation, error) {
	originalPages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse original pages: %w", err)
	}
	slicedPages, err := pagesparser.Parse(t.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sliced pages: %w", err)
	}
	for _, p := range slicedPages {
		if !lo.Contains(originalPages, p) {
			return nil, fmt.Errorf("requested sliced page %d not in original pages %v", p, originalPages)
		}
	}
	ann.Pages = t.Pages

	if err = roboflow.SlicePages(ann.RoboflowDir, slicedPages); err != nil {
		return nil, fmt.Errorf("failed to slice pages in roboflow dir: %w", err)
	}

	return ann, nil
}

// StretchCategory    string                    `json:"stretch_category"`
// Towards            string                    `json:"towards"`
// ContactType        AnnotationRuleContactType `json:"contact_type"`
// ContactSide        AnnotationRuleSide        `json:"contact_side"`
func (a *AnnotationRuleApplier) applyStretchRule(ann *model.Annotation, t *model.AnnotationRuleStretch) (*model.Annotation, error) {
	toApply, err := coco.StretchTowardsCategoryBuilder().
		Stretch(t.StretchCategory).
		Towards(t.Towards, coco.ContactTypeFromString(string(t.ContactType)), coco.SideFromString(string(t.ContactSide))).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build stretch operation from rule %+v: %w", t, err)
	}
	if err := roboflow.StretchCategoryTowardOtherCategory(ann.RoboflowDir, toApply); err != nil {
		return nil, fmt.Errorf("failed to apply stretch operation %+v to annotation %s: %w", toApply, ann.ID, err)
	}
	return ann, nil
}

func (a *AnnotationRuleApplier) applyAddMarginRule(ann *model.Annotation, t *model.AnnotationRuleAddMargin) (*model.Annotation, error) {
	toApply, err := coco.AddMarginBuilder().
		Side(coco.SideFromString(string(t.Side))).
		Margin(t.Margin).
		Category(t.Category).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build add margin operation from rule %+v: %w", t, err)
	}
	if err := roboflow.AddMargin(ann.RoboflowDir, toApply); err != nil {
		return nil, fmt.Errorf("failed to apply add margin operation %+v to annotation %s: %w", toApply, ann.ID, err)
	}
	return ann, nil
}
