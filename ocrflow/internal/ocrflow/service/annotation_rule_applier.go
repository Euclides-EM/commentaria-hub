package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/coco"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/krakenwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/roboflow"
	"github.com/samber/lo"
	"log"
)

type AnnotationRuleApplier struct {
	dataDir        string
	roboflowAPIKey string

	modelSvc *Model
}

func NewAnnotationRuleApplier(dataDir, roboflowAPIKey string, modelSvc *Model) *AnnotationRuleApplier {
	return &AnnotationRuleApplier{
		dataDir:        dataDir,
		roboflowAPIKey: roboflowAPIKey,
		modelSvc:       modelSvc,
	}
}

func (a *AnnotationRuleApplier) ApplyRules(imgPath string, ann *model.Annotation, rules []annotationrule.AnnotationRule) error {
	var err error
	log.Printf("Applying %d rules to annotation %s", len(rules), ann.ID)
	for _, rule := range rules {
		log.Printf("Applying rule %+v", rule)
		ann, err = a.ApplyRule(imgPath, ann, rule)
		if err != nil {
			return fmt.Errorf("failed to apply rule %+v: %w", rule, err)
		}
	}
	log.Printf("Applied %d rules", len(rules))
	return nil
}

func (a *AnnotationRuleApplier) ApplyRule(imgPath string, ann *model.Annotation, rule annotationrule.AnnotationRule) (*model.Annotation, error) {
	switch t := rule.(type) {
	case *annotationrule.SlicePages:
		return a.applySlicePagesRule(ann, t)
	case *annotationrule.Stretch:
		return a.applyStretchRule(ann, t)
	case *annotationrule.AddMargin:
		return a.applyAddMarginRule(ann, t)
	case *annotationrule.LinesDetect:
		return a.applyLinesDetectRule(imgPath, ann, t)
	case *annotationrule.Segment:
		return a.applySegment(imgPath, ann, t)
	default:
		log.Printf("Unknown rule type: %s", rule.GetType())
	}
	return nil, fmt.Errorf("unknown rule type: %s", rule.GetType())
}

func (a *AnnotationRuleApplier) applySlicePagesRule(ann *model.Annotation, t *annotationrule.SlicePages) (*model.Annotation, error) {
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

func (a *AnnotationRuleApplier) applyStretchRule(ann *model.Annotation, t *annotationrule.Stretch) (*model.Annotation, error) {
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

func (a *AnnotationRuleApplier) applyAddMarginRule(ann *model.Annotation, t *annotationrule.AddMargin) (*model.Annotation, error) {
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

func (a *AnnotationRuleApplier) applyLinesDetectRule(imgPath string, ann *model.Annotation, t *annotationrule.LinesDetect) (*model.Annotation, error) {
	if err := krakenwrapper.DetectLines(imgPath, ann.AltoDir, t.IncludeCategories, t.IgnoreCategories); err != nil {
		return nil, fmt.Errorf("failed to apply lines detect to annotation %s: %w", ann.ID, err)
	}
	return ann, nil
}

func (a *AnnotationRuleApplier) applySegment(imgPath string, ann *model.Annotation, t *annotationrule.Segment) (*model.Annotation, error) {
	segM, err := a.modelSvc.Get(t.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to get segmentation model %s: %w", t.Model, err)
	}

	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}
	var filenames []string
	for _, p := range pages {
		filenames = append(filenames, pagesparser.PageToPNGFilename(p))
	}

	var f func() (<-chan error, error)
	switch segM.Location {
	case model.OCRModelLocationLocal:
		ann.AltoDir = store.DatasetAnnotationAltoDir(ann, a.dataDir)
		f = func() (<-chan error, error) {
			return krakenwrapper.Recognize(
				imgPath,
				ann.AltoDir,
				segM.LocalPath,
				filenames,
			)
		}
	case model.OCRModelLocationRoboflow:
		ann.RoboflowDir = store.DatasetAnnotationRoboflowDir(ann, a.dataDir)
		f = func() (<-chan error, error) {
			return roboflow.Recognize(
				imgPath,
				ann.RoboflowDir,
				segM.Name,
				segM.Categories,
				a.roboflowAPIKey,
				filenames,
			), nil
		}
	}
	errCh, err := f()
	if err != nil {
		return nil, fmt.Errorf("failed to start annotation process: %w", err)
	}
	if recErr := <-errCh; recErr != nil {
		return nil, fmt.Errorf("failed to annotate facsimile: %w", recErr)
	}
	return ann, nil
}
