package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/annotationrules"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/krakenwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/roboflow"
	"github.com/samber/lo"
	"log"
	"os"
	"path/filepath"
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
	// rules that do not change ALTO files:
	if rule.GetType() == annotationrule.TypeSlicePages {
		return a.applySlicePagesRule(ann, rule.(*annotationrule.SlicePages))
	}

	// delete YOLO dir if exists, as it will be invalid after ALTO modification
	if ann.YoloDir != "" {
		if err := os.RemoveAll(ann.YoloDir); err != nil {
			return nil, fmt.Errorf("failed to remove YOLO dir after ALTO modification: %w", err)
		}
		ann.YoloDir = ""
	}

	// rules that modify ALTO files in a batch:
	switch t := rule.(type) {
	case *annotationrule.LinesDetect:
		return a.applyLinesDetectRule(imgPath, ann, t)
	case *annotationrule.Segment:
		return a.applySegment(imgPath, ann, t)
	}

	// rules that require per-page ALTO processing
	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}
	for _, page := range pages {
		pageImgPath := filepath.Join(imgPath, pagesparser.PageToPNGFilename(page))
		if _, err := os.Stat(pageImgPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("page image %s does not exist for annotation %s", pageImgPath, ann.ID)
		}
		pageAltoPath := filepath.Join(ann.AltoDir, pagesparser.PageToXMLFilename(page))
		if _, err := os.Stat(pageAltoPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("page ALTO %s does not exist for annotation %s", pageAltoPath, ann.ID)
		}

		af, err := alto.LoadFromFile(pageAltoPath)
		if err != nil {
			return nil, fmt.Errorf("load ALTO: %w", err)
		}

		var f func() error
		switch t := rule.(type) {
		case *annotationrule.Stretch:
			f = func() error { return a.applyStretchRule(af, t) }
		case *annotationrule.AddMargin:
			f = func() error { return a.applyAddMarginRule(af, t) }
		case *annotationrule.RemoveCategories:
			f = func() error { return a.applyRemoveCategories(af, t) }
		case *annotationrule.RemoveOverlap:
			f = func() error { return a.applyRemoveOverlap(af, t) }
		default:
			return nil, fmt.Errorf("unknown rule type: %s", rule.GetType())
		}

		if err := f(); err != nil {
			return nil, fmt.Errorf("apply rule to ALTO: %w", err)
		}

		if err := alto.SaveToFile(af, pageAltoPath); err != nil {
			return nil, fmt.Errorf("save ALTO: %w", err)
		}
	}

	return ann, nil
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

	if ann.AltoDir == "" {
		return ann, nil
	}

	des, err := os.ReadDir(ann.AltoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read alto dir: %w", err)
	}
	for _, de := range des {
		if de.IsDir() || filepath.Ext(de.Name()) != ".xml" {
			continue
		}
		pageNum, err := pagesparser.FileNameToPage(de.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to get page number from filename %s: %w", de.Name(), err)
		}
		if !lo.Contains(slicedPages, pageNum) {
			if err := os.Remove(filepath.Join(ann.AltoDir, de.Name())); err != nil {
				return nil, fmt.Errorf("failed to remove alto file %s: %w", de.Name(), err)
			}
		}
	}

	return ann, nil
}

func (a *AnnotationRuleApplier) applyStretchRule(af *alto.Alto, t *annotationrule.Stretch) error {
	toApply, err := annotationrules.StretchTowardsCategoryBuilder().
		Stretch(t.StretchCategory).
		Towards(t.Towards, annotationrules.ContactTypeFromString(string(t.ContactType)), annotationrules.SideFromString(string(t.ContactSide))).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build stretch operation from rule %+v: %w", t, err)
	}
	if err := alto.ApplyStretchTowardsCategoryALTO(af, toApply); err != nil {
		return fmt.Errorf("failed to apply stretch operation %+v: %w", toApply, err)
	}
	return nil
}

func (a *AnnotationRuleApplier) applyAddMarginRule(af *alto.Alto, t *annotationrule.AddMargin) error {
	toApply, err := annotationrules.AddMarginBuilder().
		Side(annotationrules.SideFromString(string(t.Side))).
		Margin(t.Margin).
		Category(t.Category).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build add margin operation from rule %+v: %w", t, err)
	}
	if err := alto.ApplyAddMarginALTO(af, toApply); err != nil {
		return fmt.Errorf("failed to apply add margin operation %+v: %w", toApply, err)
	}
	return nil
}

func (a *AnnotationRuleApplier) applyRemoveCategories(af *alto.Alto, t *annotationrule.RemoveCategories) error {
	if err := alto.ApplyRemoveCategoriesALTO(af, t.Categories); err != nil {
		return fmt.Errorf("failed to remove categories %v: %w", t.Categories, err)
	}
	return nil
}

func (a *AnnotationRuleApplier) applyRemoveOverlap(af *alto.Alto, t *annotationrule.RemoveOverlap) error {
	if err := alto.FixNoOverlap(af, t.Categories, t.Precision); err != nil {
		return fmt.Errorf("failed to apply remove overlap operation %+v: %w", t, err)
	}
	return nil
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
		ann.AltoDir = store.DatasetAnnotationAltoDir(ann, a.dataDir)
		f = func() (<-chan error, error) {
			return roboflow.Recognize(
				imgPath,
				ann.AltoDir,
				segM.Name,
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
