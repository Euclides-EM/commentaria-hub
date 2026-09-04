package service

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotationrule"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/annotationrules"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/krakenwrapper"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/roboflow"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/transcriptioncorrector"
	"github.com/samber/lo"
)

type AnnotationRuleApplier struct {
	modelSvc         *Model
	remoteDetectSvc  *AnnotationDetectionRemote
	fileSysMgt       *filesys.Manager
	roboflowAPIKey   string
	annotationStore  annotationRuleAnnotationStore
	datasetStore     annotationRuleDatasetStore
	transcriptionLLM transcriptioncorrector.Executor
}

type annotationRuleAnnotationStore interface {
	GetAnnotation(datasetID, id string) (*annotation.Annotation, error)
}

type annotationRuleDatasetStore interface {
	GetDataset(id string) (*model.Dataset, error)
}

func NewAnnotationRuleApplier(modelSvc *Model, fileSysMgt *filesys.Manager, roboflowAPIKey string, remoteDetect *AnnotationDetectionRemote, annotationStore annotationRuleAnnotationStore, datasetStore annotationRuleDatasetStore, transcriptionLLM transcriptioncorrector.Executor) *AnnotationRuleApplier {
	return &AnnotationRuleApplier{
		modelSvc:         modelSvc,
		remoteDetectSvc:  remoteDetect,
		fileSysMgt:       fileSysMgt,
		roboflowAPIKey:   roboflowAPIKey,
		annotationStore:  annotationStore,
		datasetStore:     datasetStore,
		transcriptionLLM: transcriptionLLM,
	}
}

func (a *AnnotationRuleApplier) ApplyRules(imgPath string, ann *annotation.Annotation, rules []annotationrule.AnnotationRule) error {
	return a.ApplyRulesWithRemoteProgress(imgPath, ann, rules, nil)
}

func (a *AnnotationRuleApplier) ApplyRulesWithRemoteProgress(imgPath string, ann *annotation.Annotation, rules []annotationrule.AnnotationRule, onSubmitted func(string)) error {
	var err error
	log.Printf("Applying %d rules to annotation %s", len(rules), ann.ID)
	hasRemoteWork := false
	for _, rule := range rules {
		log.Printf("Applying rule %+v", rule)
		switch t := rule.(type) {
		case *annotationrule.ModelDetect:
			hasRemoteWork = hasRemoteWork || t.UseGPUFarm
		case *annotationrule.LinesDetect:
			hasRemoteWork = hasRemoteWork || t.UseGPUFarm
		}
		ann, err = a.applyRule(imgPath, ann, rule, onSubmitted)
		if err != nil {
			return fmt.Errorf("failed to apply rule %+v: %w", rule, err)
		}
	}
	if hasRemoteWork {
		log.Printf("Dispatched %d rules for annotation %s; GPU farm processing continues asynchronously until its result upload completes", len(rules), ann.ID)
	} else {
		log.Printf("Applied %d rules to annotation %s", len(rules), ann.ID)
	}
	return nil
}

func (a *AnnotationRuleApplier) ApplyRule(imgPath string, ann *annotation.Annotation, rule annotationrule.AnnotationRule) (*annotation.Annotation, error) {
	return a.applyRule(imgPath, ann, rule, nil)
}

func (a *AnnotationRuleApplier) applyRule(imgPath string, ann *annotation.Annotation, rule annotationrule.AnnotationRule, onSubmitted func(string)) (*annotation.Annotation, error) {
	if t, ok := rule.(*annotationrule.LLMTranscriptionCorrector); ok {
		return a.applyLLMTranscriptionCorrector(imgPath, ann, t)
	}

	// delete YOLO dir if exists, as it will be invalid after ALTO modification
	if err := os.RemoveAll(a.fileSysMgt.DatasetAnnotationYoloDir(ann)); err != nil {
		return nil, fmt.Errorf("failed to remove YOLO dir after ALTO modification: %w", err)
	}

	// rules that modify ALTO files in a batch:
	switch t := rule.(type) {
	case *annotationrule.SlicePages:
		return a.applySlicePagesRule(ann, t)
	case *annotationrule.LinesDetect:
		return a.applyLinesDetectRule(imgPath, ann, t, onSubmitted)
	case *annotationrule.ModelDetect:
		return a.applyModelDetect(imgPath, ann, t, onSubmitted)
	case *annotationrule.TextBlockCorrections:
		return a.applyTextBlockCorrection(ann, t)
	}

	// rules that require per-page ALTO processing
	pages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}
	for _, page := range pages {
		if err := a.fileSysMgt.ApplyToAltoPage(ann, page, func(af *alto.Alto) error {
			var f func() error
			switch t := rule.(type) {
			case *annotationrule.Stretch:
				f = func() error { return a.applyStretchRule(af, t) }
			case *annotationrule.AddMargin:
				f = func() error { return a.applyAddMarginRule(af, t) }
			case *annotationrule.RemoveCategories:
				f = func() error { return a.applyRemoveCategories(af, t) }
			case *annotationrule.RenameCategories:
				f = func() error { return a.applyRenameCategories(af, t) }
			case *annotationrule.RemoveOverlap:
				f = func() error { return a.applyRemoveOverlap(af, t) }
			case *annotationrule.ResolveOverlapWithPriority:
				f = func() error { return a.applyResolveOverlapWithPriority(af, t) }
			case *annotationrule.RecategorizeByAlignment:
				f = func() error { return a.applyRecategorizeByAlignment(af, t) }
			case *annotationrule.LimitCategoryZones:
				f = func() error { return a.applyLimitCategoryZones(af, t) }
			case *annotationrule.ReassignTextLinesByTolerance:
				f = func() error { return a.applyReassignTextLinesByTolerance(af, t) }
			default:
				return fmt.Errorf("unknown rule type: %s", rule.GetType())
			}

			if err := f(); err != nil {
				return fmt.Errorf("apply rule to ALTO: %w", err)
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("failed to apply rule to page %d: %w", page, err)
		}
	}

	return ann, nil
}

func (a *AnnotationRuleApplier) applyLLMTranscriptionCorrector(imgPath string, ann *annotation.Annotation, rule *annotationrule.LLMTranscriptionCorrector) (*annotation.Annotation, error) {
	if !ann.Ocred {
		return nil, fmt.Errorf("LLM transcription corrector is only applicable to OCR-ed annotations; annotation %s is not OCR-ed", ann.ID)
	}
	if strings.TrimSpace(rule.Provider) == "" {
		return nil, fmt.Errorf("LLM transcription corrector provider is required")
	}
	if strings.TrimSpace(rule.Model) == "" {
		return nil, fmt.Errorf("LLM transcription corrector model is required")
	}
	if a.transcriptionLLM == nil {
		return nil, fmt.Errorf("LLM transcription corrector client is not configured")
	}

	altoDirs := []string{a.fileSysMgt.DatasetAnnotationAltoDir(ann)}
	seen := map[string]struct{}{ann.ID: {}}
	if len(rule.AdditionalAnnotations) > 0 && a.annotationStore == nil {
		return nil, fmt.Errorf("annotation store is not configured for additional_annotations")
	}
	for _, additionalID := range rule.AdditionalAnnotations {
		additionalID = strings.TrimSpace(additionalID)
		if additionalID == "" {
			continue
		}
		if _, exists := seen[additionalID]; exists {
			return nil, fmt.Errorf("additional annotation %s is duplicated or is the target annotation", additionalID)
		}
		seen[additionalID] = struct{}{}
		additional, err := a.annotationStore.GetAnnotation(ann.DatasetID, additionalID)
		if err != nil {
			return nil, fmt.Errorf("get additional annotation %s: %w", additionalID, err)
		}
		if additional == nil {
			return nil, fmt.Errorf("additional annotation %s was not found in dataset %s", additionalID, ann.DatasetID)
		}
		if !additional.Ocred {
			return nil, fmt.Errorf("additional annotation %s is not OCR-ed", additionalID)
		}
		altoDirs = append(altoDirs, a.fileSysMgt.DatasetAnnotationAltoDir(additional))
	}

	var transcriptionDirs []string
	if rule.IncludeEditionTranscription {
		if a.datasetStore == nil {
			return nil, fmt.Errorf("dataset store is not configured for include_edition_transcription")
		}
		dataset, err := a.datasetStore.GetDataset(ann.DatasetID)
		if err != nil {
			return nil, fmt.Errorf("get dataset %s for edition transcription: %w", ann.DatasetID, err)
		}
		if dataset == nil {
			return nil, fmt.Errorf("dataset %s not found", ann.DatasetID)
		}
		if strings.TrimSpace(dataset.EditionID) == "" {
			return nil, fmt.Errorf("dataset %s has no edition for include_edition_transcription", ann.DatasetID)
		}
		transcriptionDirs = append(transcriptionDirs, a.fileSysMgt.EditionTxtTranscriptionDir(dataset.EditionID))
	}

	pageNumbers, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("parse pages for annotation %s: %w", ann.ID, err)
	}
	if strings.TrimSpace(rule.Pages) != "" {
		requestedPages, err := pagesparser.IntRange(rule.Pages)
		if err != nil {
			return nil, fmt.Errorf("parse LLM transcription corrector pages: %w", err)
		}
		annotationPages := make(map[int]struct{}, len(pageNumbers))
		for _, pageNumber := range pageNumbers {
			annotationPages[pageNumber] = struct{}{}
		}
		for _, pageNumber := range requestedPages {
			if _, ok := annotationPages[pageNumber]; !ok {
				return nil, fmt.Errorf("LLM transcription corrector page %d is not in annotation %s", pageNumber, ann.ID)
			}
		}
		pageNumbers = requestedPages
	}
	pageKeys := make([]string, 0, len(pageNumbers))
	for _, pageNumber := range pageNumbers {
		pageKeys = append(pageKeys, pagesparser.PageToFilename(pageNumber, ""))
	}

	cfg := transcriptioncorrector.Config{
		ALTODirs:          altoDirs,
		TranscriptionDirs: transcriptionDirs,
		PageKeys:          pageKeys,
		ImagesDir:         imgPath,
		OutputDir:         a.fileSysMgt.AnnotationTxtTranscriptionDir(ann),
		Rounds:            rule.Rounds,
		SkipExisting:      rule.SkipExisting,
		Provider:          strings.TrimSpace(rule.Provider),
		Model:             strings.TrimSpace(rule.Model),
	}
	usage, err := transcriptioncorrector.Run(cfg, a.transcriptionLLM)
	if err != nil {
		return nil, fmt.Errorf("run LLM transcription corrector for annotation %s: %w", ann.ID, err)
	}
	rule.Usage = &usage
	return ann, nil
}

func (a *AnnotationRuleApplier) applySlicePagesRule(ann *annotation.Annotation, t *annotationrule.SlicePages) (*annotation.Annotation, error) {
	originalPages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse original pages: %w", err)
	}

	if t.Pages == "" && t.RandomPages > 0 {
		selectPages := make([]int, len(originalPages))
		copy(selectPages, originalPages)
		if len(selectPages) < t.RandomPages {
			return nil, fmt.Errorf("requested random pages %d exceeds total pages %d", t.RandomPages, len(originalPages))
		}
		rand.Shuffle(len(selectPages), func(i, j int) {
			selectPages[i], selectPages[j] = selectPages[j], selectPages[i]
		})
		t.Pages = pagesparser.ToString(selectPages[:t.RandomPages])
		t.RandomPages = 0
	}

	slicedPages, err := pagesparser.IntRange(t.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sliced pages: %w", err)
	}
	for _, p := range slicedPages {
		if !lo.Contains(originalPages, p) {
			return nil, fmt.Errorf("requested sliced page %d not in original pages %v", p, originalPages)
		}
	}
	ann.Pages = t.Pages

	if !ann.Segmented {
		return ann, nil
	}

	des, err := os.ReadDir(a.fileSysMgt.DatasetAnnotationAltoDir(ann))
	if err != nil {
		return nil, fmt.Errorf("failed to read alto dir: %w", err)
	}
	for _, de := range des {
		if de.IsDir() || filepath.Ext(de.Name()) != ".xml" || filepath.Base(de.Name()) == "METS.xml" {
			continue
		}
		pageNum, err := pagesparser.FileNameToPage(de.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to get page number from filename %s: %w", de.Name(), err)
		}
		if !lo.Contains(slicedPages, pageNum) {
			if err := os.Remove(filepath.Join(a.fileSysMgt.DatasetAnnotationAltoDir(ann), de.Name())); err != nil {
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

func (a *AnnotationRuleApplier) applyTextBlockCorrection(ann *annotation.Annotation, t *annotationrule.TextBlockCorrections) (*annotation.Annotation, error) {
	correctionByPage := make(map[int][]*annotationrule.TextBlockCorrection)
	for _, c := range t.Corrections {
		correctionByPage[c.Page] = append(correctionByPage[c.Page], c)
	}

	pages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}

	for _, page := range pages {
		corrections, ok := correctionByPage[page]
		if !ok {
			continue
		}
		if err := a.fileSysMgt.ApplyToAltoPage(ann, page, func(af *alto.Alto) error {
			for _, c := range corrections {
				if err := alto.ApplyTextBlockCorrectionALTO(af, c.TextBlockID, c.Old, c.Correction); err != nil {
					return fmt.Errorf("failed to apply text block correction %+v: %w", c, err)
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("failed to apply text block corrections to page %d: %w", page, err)
		}
	}
	return ann, nil
}

func (a *AnnotationRuleApplier) applyRemoveCategories(af *alto.Alto, t *annotationrule.RemoveCategories) error {
	if err := alto.ApplyRemoveCategoriesALTO(af, t.Categories); err != nil {
		return fmt.Errorf("failed to remove categories %v: %w", t.Categories, err)
	}
	return nil
}

func (a *AnnotationRuleApplier) applyRenameCategories(af *alto.Alto, t *annotationrule.RenameCategories) error {
	if err := alto.ApplyRenameCategoriesALTO(af, t.Renames); err != nil {
		return fmt.Errorf("failed to rename categories %v: %w", t.Renames, err)
	}
	return nil
}

func (a *AnnotationRuleApplier) applyRemoveOverlap(af *alto.Alto, t *annotationrule.RemoveOverlap) error {
	if err := alto.FixNoOverlap(af, t.Categories, t.Precision); err != nil {
		return fmt.Errorf("failed to apply remove overlap operation %+v: %w", t, err)
	}
	return nil
}

func (a *AnnotationRuleApplier) applyResolveOverlapWithPriority(af *alto.Alto, t *annotationrule.ResolveOverlapWithPriority) error {
	if err := alto.ResolveOverlapWithPriority(af, t.DominantCategory, t.SuppressedCategory, t.MinOverlap); err != nil {
		return fmt.Errorf("failed to apply resolve overlap with priority %+v: %w", t, err)
	}
	return nil
}

func (a *AnnotationRuleApplier) applyRecategorizeByAlignment(af *alto.Alto, t *annotationrule.RecategorizeByAlignment) error {
	if err := alto.RecategorizeByAlignment(af, t.OriginalCategory, t.TargetCategory, t.RelativeTo.Category, string(t.RelativeTo.Alignment), t.TolerancePx); err != nil {
		return fmt.Errorf("failed to apply recategorize by alignment %+v: %w", t, err)
	}
	return nil
}

func (a *AnnotationRuleApplier) applyLimitCategoryZones(af *alto.Alto, t *annotationrule.LimitCategoryZones) error {
	if err := alto.LimitCategoryZones(af, t.Category, t.MaxCount, string(t.KeepPosition)); err != nil {
		return fmt.Errorf("failed to apply limit category zones %+v: %w", t, err)
	}
	return nil
}

func (a *AnnotationRuleApplier) applyReassignTextLinesByTolerance(af *alto.Alto, t *annotationrule.ReassignTextLinesByTolerance) error {
	moved, err := alto.ReassignTextLinesByTolerance(af, t.FromCategory, t.ToCategory, t.PrecisionPx, t.MinOverlap)
	if err != nil {
		return fmt.Errorf("failed to apply reassign text lines by tolerance %+v: %w", t, err)
	}
	log.Printf("[DEBUG] Reassigned %d text lines from category %s to %s", moved, t.FromCategory, t.ToCategory)
	return nil
}

func (a *AnnotationRuleApplier) applyLinesDetectRule(imgPath string, ann *annotation.Annotation, t *annotationrule.LinesDetect, onSubmitted func(string)) (*annotation.Annotation, error) {
	if t.UseGPUFarm {
		return a.applyLinesDetectRuleRemote(imgPath, ann, t, onSubmitted)
	}
	if err := krakenwrapper.DetectLines(imgPath, a.fileSysMgt.DatasetAnnotationAltoDir(ann), t.IncludeCategories, t.IgnoreCategories); err != nil {
		return nil, fmt.Errorf("failed to apply lines detect to annotation %s: %w", ann.ID, err)
	}
	ann.LinesDetected = true
	return ann, nil
}

func (a *AnnotationRuleApplier) applyModelDetect(imgPath string, ann *annotation.Annotation, t *annotationrule.ModelDetect, onSubmitted func(string)) (*annotation.Annotation, error) {
	m, err := a.modelSvc.Get(t.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to get segmentation model %s: %w", t.Model, err)
	}

	if m.Type != t.ModelType {
		return nil, fmt.Errorf("model type mismatch for model %s: the model is of type %s but rule requires type %s", t.Model, m.Type, t.ModelType)
	}
	if m.Type == common.OCRModelTypeOCR && !ann.LinesDetected {
		return nil, fmt.Errorf("cannot apply OCR model detect rule to annotation %s before text lines are detected", ann.ID)
	}

	pages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}

	if t.UseGPUFarm {
		return a.applyModelDetectRemote(imgPath, ann, m, pages, onSubmitted)
	}

	var filenames []string
	for _, p := range pages {
		filenames = append(filenames, pagesparser.PageToPNGFilename(p))
	}

	var f func() (<-chan error, error)
	switch {
	case m.Type == common.OCRModelTypeOCR:
		f = func() (<-chan error, error) {
			var imgAndAltoPaths [][2]string
			for _, p := range pages {
				pathToImg := filepath.Join(imgPath, pagesparser.PageToPNGFilename(p))
				pathToImg, err = filepath.Abs(pathToImg)
				if err != nil {
					return nil, fmt.Errorf("could not determine absolute path of input image %s: %w", pathToImg, err)
				}
				pathToAlto := filepath.Join(a.fileSysMgt.DatasetAnnotationAltoDir(ann), pagesparser.PageToXMLFilename(p))
				pathToAlto, err = filepath.Abs(pathToAlto)
				if err != nil {
					return nil, fmt.Errorf("could not determine absolute path of input ALTO %s: %w", pathToAlto, err)
				}
				imgAndAltoPaths = append(imgAndAltoPaths, [2]string{
					pathToImg,
					pathToAlto,
				})
			}
			return krakenwrapper.RecognizeTextWithMapping(imgAndAltoPaths, a.fileSysMgt.ModelPath(m))
		}
	case m.Type == common.OCRModelTypeSegment && m.Location == model.OCRModelLocationLocal:
		f = func() (<-chan error, error) {
			return krakenwrapper.Segment(
				imgPath,
				a.fileSysMgt.DatasetAnnotationAltoDir(ann),
				a.fileSysMgt.ModelPath(m),
				filenames,
			)
		}
	case m.Type == common.OCRModelTypeSegment && m.Location == model.OCRModelLocationRoboflow:
		f = func() (<-chan error, error) {
			return roboflow.Recognize(
				imgPath,
				a.fileSysMgt.DatasetAnnotationAltoDir(ann),
				m.Name,
				a.roboflowAPIKey,
				filenames,
			), nil
		}
	}
	errCh, err := f()
	if err != nil {
		return nil, fmt.Errorf("failed to start detection using model %s: %w", m.ID, err)
	}
	if recErr := <-errCh; recErr != nil {
		return nil, fmt.Errorf("failed to apply model detect to annotation %s: %w", ann.ID, recErr)
	}
	switch m.Type {
	case common.OCRModelTypeOCR:
		ann.Ocred = true
	case common.OCRModelTypeSegment:
		ann.Segmented = true
	}
	return ann, nil
}
