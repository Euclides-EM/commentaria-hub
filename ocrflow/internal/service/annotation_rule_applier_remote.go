package service

import (
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
)

func (a *AnnotationRuleApplier) applyLinesDetectRuleRemote(imgPath string, ann *annotation.Annotation, t *annotationrule.LinesDetect) (*annotation.Annotation, error) {
	pages, err := parseAnnotationPages(ann)
	if err != nil {
		return nil, err
	}
	if err := a.submitGPUFarmDetection(remoteDetectionRequest{
		Mode:              annotation.DetectionModeLines,
		ImageDir:          imgPath,
		Annotation:        ann,
		Pages:             pages,
		IncludeCategories: t.IncludeCategories,
		IgnoreCategories:  t.IgnoreCategories,
	}); err != nil {
		return nil, fmt.Errorf("failed to submit lines detect for annotation %s to GPU farm: %w", ann.ID, err)
	}
	return ann, nil
}

func (a *AnnotationRuleApplier) applyModelDetectRemote(imgPath string, ann *annotation.Annotation, m *model.Model, pages []int) (*annotation.Annotation, error) {
	if m.Location != model.OCRModelLocationLocal {
		return nil, fmt.Errorf("GPU farm detection requires a local model, got %s", m.Location)
	}
	if err := a.submitGPUFarmDetection(remoteDetectionRequest{
		Mode:       annotation.DetectionModeForModelType(m.Type),
		ImageDir:   imgPath,
		Annotation: ann,
		Pages:      pages,
		Model:      m,
		ModelPath:  a.fileSysMgt.ModelPath(m),
	}); err != nil {
		return nil, fmt.Errorf("failed to submit model detect for annotation %s to GPU farm: %w", ann.ID, err)
	}
	return ann, nil
}

func (a *AnnotationRuleApplier) submitGPUFarmDetection(req remoteDetectionRequest) error {
	if a.remoteDetectSvc == nil {
		return fmt.Errorf("GPU farm detection is not configured")
	}
	return a.remoteDetectSvc.Submit(req)
}

func parseAnnotationPages(ann *annotation.Annotation) ([]int, error) {
	pages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}
	return pages, nil
}
