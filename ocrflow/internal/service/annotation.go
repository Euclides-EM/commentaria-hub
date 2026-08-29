package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotationrule"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/titlepage"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/formatcov"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/idgen"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/markdown"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/name"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
	"github.com/samber/lo"
	"github.com/tiendc/go-deepcopy"
)

// todo: add interfaces to all services

type Annotation struct {
	datasetSvc        *Dataset
	datasetImgSvc     *DatasetImg
	ruleApplier       *AnnotationRuleApplier
	featureResultsSvc *Result
	fileSysMgt        *filesys.Manager
	annotationStore   *store.AnnotationSQL
	activeRuleRuns    sync.Map
}

var ErrAnnotationRuleRunInProgress = errors.New("annotation rule application already in progress")

func NewAnnotationsService(datasetSvc *Dataset, datasetImgSvc *DatasetImg, ruleApplier *AnnotationRuleApplier, featureResultsSvc *Result, fileSysMgt *filesys.Manager, annotationStore *store.AnnotationSQL) *Annotation {
	return &Annotation{
		datasetSvc:        datasetSvc,
		datasetImgSvc:     datasetImgSvc,
		ruleApplier:       ruleApplier,
		featureResultsSvc: featureResultsSvc,
		fileSysMgt:        fileSysMgt,
		annotationStore:   annotationStore,
	}
}

func (a *Annotation) ListAnnotations(id string) ([]*annotation.Annotation, error) {
	annotations, err := a.annotationStore.ListAnnotationsByDatasetID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations from store: %w", err)
	}
	if err := a.populateTranscriptionFallbacks(annotations); err != nil {
		return nil, err
	}
	return annotations, nil
}

func (a *Annotation) Get(datasetId, id string) (*annotation.Annotation, error) {
	ann, err := a.annotationStore.GetAnnotation(datasetId, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation from store: %w", err)
	}
	if ann == nil || ann.DatasetID != datasetId {
		return nil, fmt.Errorf("annotation not found")
	}
	if err := a.populateTranscriptionFallbacks([]*annotation.Annotation{ann}); err != nil {
		return nil, err
	}
	return ann, nil
}

func (a *Annotation) populateTranscriptionFallbacks(annotations []*annotation.Annotation) error {
	fallbackFormatByDatasetID := make(map[string]*annotation.TranscriptionFallback)
	for _, ann := range annotations {
		if ann.DatasetID == "" {
			continue
		}
		fallback, ok := fallbackFormatByDatasetID[ann.DatasetID]
		if !ok {
			var err error
			fallback, err = a.transcriptionFallback(ann)
			if err != nil {
				return err
			}
			fallbackFormatByDatasetID[ann.DatasetID] = fallback
		}
		if fallback != nil {
			ann.TranscriptionFallback = fallback
		}
	}
	return nil
}

func (a *Annotation) transcriptionFallback(ann *annotation.Annotation) (*annotation.TranscriptionFallback, error) {
	ds, err := a.datasetSvc.Get(ann.DatasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset from store: %w", err)
	}

	annPages, err := pagesparser.Range(ann.Pages)
	if err != nil {
		return nil, err
	}

	if lo.EveryBy(annPages, func(page string) bool {
		_, _, err := a.fileSysMgt.RetrieveAnnotationAltoPage(ann, page)
		return err == nil
	}) {
		return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatALTO, annotation.TranscriptionLevelAnnotation, false), nil
	}
	if lo.SomeBy(annPages, func(page string) bool {
		_, _, err := a.fileSysMgt.RetrieveAnnotationAltoPage(ann, page)
		return err == nil
	}) {
		return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatALTO, annotation.TranscriptionLevelAnnotation, true), nil
	}

	if lo.EveryBy(annPages, func(page string) bool {
		_, err := a.fileSysMgt.RetrieveAnnotationMarkdownPage(ann, page)
		return err == nil
	}) {
		return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatMarkdown, annotation.TranscriptionLevelAnnotation, false), nil
	}
	if lo.SomeBy(annPages, func(page string) bool {
		_, err := a.fileSysMgt.RetrieveAnnotationMarkdownPage(ann, page)
		return err == nil
	}) {
		return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatMarkdown, annotation.TranscriptionLevelAnnotation, true), nil
	}

	if lo.EveryBy(annPages, func(page string) bool {
		_, _, err := a.fileSysMgt.RetrieveEditionTXTPage(ds.EditionID, page)
		return err == nil
	}) {
		return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatText, annotation.TranscriptionLevelAnnotation, false), nil
	}
	if lo.SomeBy(annPages, func(page string) bool {
		_, _, err := a.fileSysMgt.RetrieveEditionTXTPage(ds.EditionID, page)
		return err == nil
	}) {
		return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatText, annotation.TranscriptionLevelAnnotation, true), nil
	}

	if ds.EditionID != "" {
		edPages, err := pagesparser.IntRange(ann.Pages)
		if err != nil {
			return nil, err
		}

		if lo.EveryBy(edPages, func(page int) bool {
			_, _, err := a.fileSysMgt.RetrieveEditionAltoPage(ds.EditionID, page)
			return err == nil
		}) {
			return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatALTO, annotation.TranscriptionLevelDataset, false), nil
		}
		if lo.SomeBy(edPages, func(page int) bool {
			_, _, err := a.fileSysMgt.RetrieveEditionAltoPage(ds.EditionID, page)
			return err == nil
		}) {
			return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatALTO, annotation.TranscriptionLevelDataset, true), nil
		}

		if lo.EveryBy(edPages, func(page int) bool {
			_, err := a.fileSysMgt.RetrieveEditionMarkdownPage(ds.EditionID, page)
			return err == nil
		}) {
			return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatMarkdown, annotation.TranscriptionLevelDataset, false), nil
		}
		if lo.SomeBy(edPages, func(page int) bool {
			_, err := a.fileSysMgt.RetrieveEditionMarkdownPage(ds.EditionID, page)
			return err == nil
		}) {
			return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatMarkdown, annotation.TranscriptionLevelDataset, true), nil
		}

		if lo.EveryBy(edPages, func(page int) bool {
			_, _, err := a.fileSysMgt.RetrieveEditionTXTPage(ds.EditionID, strconv.Itoa(page))
			return err == nil
		}) {
			return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatText, annotation.TranscriptionLevelDataset, false), nil
		}
		if lo.SomeBy(edPages, func(page int) bool {
			_, _, err := a.fileSysMgt.RetrieveEditionTXTPage(ds.EditionID, strconv.Itoa(page))
			return err == nil
		}) {
			return annotation.NewTranscriptionFallback(annotation.TranscriptionFormatText, annotation.TranscriptionLevelDataset, true), nil
		}
	}

	return nil, nil
}

var errStopTranscriptionFallbackScan = errors.New("stop transcription fallback scan")

func transcriptionFallbackFormatInDir(dir string) (string, error) {
	format := ""
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch filepath.Base(path) {
		case "original.xml":
			format = "alto"
			return errStopTranscriptionFallbackScan
		case "original.md":
			if format == "" || format == "text" {
				format = "markdown"
			}
		case "original.txt":
			if format == "" {
				format = "text"
			}
		}
		return nil
	})
	if errors.Is(err, errStopTranscriptionFallbackScan) {
		return format, nil
	}
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("scan transcription fallback dir %s: %w", dir, err)
	}
	return format, nil
}

func (a *Annotation) Create(datasetID string, ann *annotation.Annotation, copyFeatureResults bool) (*annotation.Annotation, error) {
	// validate dataset exists
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	imgPath := a.fileSysMgt.DatasetImagesDir(ds)
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no images found for dataset %s", datasetID)
	}
	anns, err := a.ListAnnotations(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations for dataset: %w", err)
	}

	// assign basic fields
	ann.ID = idgen.GenerateID(store.AnnotationIDPrefix)
	ann.DatasetID = datasetID
	ann.Name = name.NextAvailable(lo.Map(anns, func(a *annotation.Annotation, _ int) string { return a.Name }), ann.Name)

	if ann.Pages == "" {
		// infer pages from existing images
		pages, err := store.InferPages(imgPath, annotation.FormatAlto)
		if err != nil {
			return nil, fmt.Errorf("failed to infer pages from existing dataset images: %w", err)
		}
		ann.Pages = pagesparser.ToString(pages)
	}

	// verify page images exist for all specified pages
	pages, err := pagesparser.Range(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages: %w", err)
	}

	for _, p := range pages {
		if datasetID == titlepage.DatasetID {
			if _, err := a.datasetImgSvc.GetImageMetadata(datasetID, p); err != nil {
				log.Printf("[WARN] failed to fetch image metadata for dataset %s: %v", datasetID, err)
			}
		} else {
			filename := pagesparser.PageOrKeyToPNGFilename(p)
			if _, err := os.Stat(path.Join(imgPath, filename)); err != nil {
				return nil, fmt.Errorf("no such file %s in existing dataset", filename)
			}
		}
	}
	var origAnn *annotation.Annotation
	if ann.OriginAnnotationID != "" {
		origAnn, err = a.Get(datasetID, ann.OriginAnnotationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get origin annotation from store: %w", err)
		}
		ann.AppliedRules = origAnn.AppliedRules
		ann.LinesDetected = origAnn.LinesDetected
		ann.Ocred = origAnn.Ocred
		ann.Segmented = origAnn.Segmented
		ann.Pages = origAnn.Pages
	}
	if err := a.annotationStore.InsertAnnotation(ann); err != nil {
		return nil, fmt.Errorf("failed to insert annotation to store: %w", err)
	}
	if origAnn != nil && origAnn.Segmented {
		if err := futils.CopyDir(a.fileSysMgt.DatasetAnnotationAltoDir(origAnn), a.fileSysMgt.DatasetAnnotationAltoDir(ann)); err != nil {
			return nil, fmt.Errorf("failed to copy annotations for new annotation: %w", err)
		}
	}
	if copyFeatureResults && ann.OriginAnnotationID != "" {
		if err := a.featureResultsSvc.CopyResults(ann.DatasetID, ann.OriginAnnotationID, ann.DatasetID, ann.ID); err != nil {
			return nil, fmt.Errorf("failed to copy feature results for new annotation: %w", err)
		}
	}
	return ann, nil
}

func (a *Annotation) CreateFromZip(aum *annotation.UploadMetadata, save func(dstPath string) error) (*annotation.Annotation, error) {
	_, err := a.datasetSvc.Get(aum.DatasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	ann := &annotation.Annotation{
		Meta:               common.NewMeta(idgen.GenerateID(store.AnnotationIDPrefix)).WithName(aum.Name).WithDescription(aum.Description),
		DatasetID:          aum.DatasetID,
		Segmented:          aum.Segmented,
		GroundTruth:        aum.GroundTruth,
		Ocred:              aum.Ocred,
		LinesDetected:      aum.LinesDetected,
		Hidden:             aum.Hidden,
		OriginAnnotationID: aum.OriginAnnotationID,
	}
	if aum.OriginAnnotationID != "" {
		originAnn, err := a.annotationStore.GetAnnotation(aum.DatasetID, aum.OriginAnnotationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get origin annotation from store: %w", err)
		}
		if originAnn != nil {
			ann.AppliedRules = originAnn.AppliedRules
		}
	}
	if aum.Segmented && aum.SegmentModelID != "" {
		ann.AppliedRules = append(ann.AppliedRules, annotationrule.NewModelDetect(aum.SegmentModelID))
	}
	if aum.Ocred && aum.OCRModelID != "" {
		ann.AppliedRules = append(ann.AppliedRules, annotationrule.NewOCRModelDetect(aum.OCRModelID))
	}
	dstPath := a.fileSysMgt.DatasetAnnotationAltoDir(ann)
	if aum.Format == annotation.FormatYolo {
		dstPath = a.fileSysMgt.DatasetAnnotationYoloDir(ann)
	}
	if err := save(dstPath); err != nil {
		return nil, fmt.Errorf("failed to store uploaded annotations: %w", err)
	}
	if aum.Format == annotation.FormatYolo {
		if err := formatcov.Yolo2Alto(a.fileSysMgt.DatasetAnnotationYoloDir(ann), a.fileSysMgt.DatasetAnnotationAltoDir(ann)); err != nil {
			return nil, fmt.Errorf("failed to convert YOLO annotations to ALTO: %w", err)
		}
	}

	pages, err := store.InferPages(a.fileSysMgt.DatasetAnnotationAltoDir(ann), annotation.FormatAlto)
	if err != nil {
		return nil, fmt.Errorf("failed to infer pages from uploaded annotations: %w", err)
	}
	ann.Pages = pagesparser.ToString(pages)

	if ann.GroundTruth {
		for _, p := range pages {
			if err := a.fileSysMgt.ApplyToAltoPage(ann, int(p), func(a *alto.Alto) error {
				alto.ApplyFullCertainty(a)
				return nil
			}); err != nil {
				return nil, fmt.Errorf("apply full certainty to ALTO: %w", err)
			}
		}
	}

	if err := a.annotationStore.InsertAnnotation(ann); err != nil {
		return nil, fmt.Errorf("failed to insert annotation to store: %w", err)
	}
	return ann, nil
}

func (a *Annotation) PrepareApplyRules(datasetID string, id string, aar *annotationrule.ApplyRules) (*annotation.Annotation, error) {
	ann, err := a.Get(datasetID, id)
	if err != nil {
		return nil, err
	}

	if aar.Action != annotationrule.ApplyRulesActionCreateNew {
		return ann, nil
	}

	ann, err = a.Duplicate(datasetID, id, aar.Name, aar.Description, aar.CopyFeatureResults)
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate annotation for applying rules: %w", err)
	}
	ann.GroundTruth = false
	if err := a.annotationStore.UpdateAnnotation(ann); err != nil {
		return nil, fmt.Errorf("failed to update duplicated annotation in store: %w", err)
	}

	return ann, nil
}

func (a *Annotation) ExecuteApplyRules(datasetID string, id string, aar *annotationrule.ApplyRules) (*annotation.Annotation, error) {
	return a.ExecuteApplyRulesWithRemoteProgress(datasetID, id, aar, nil)
}

func (a *Annotation) ExecuteApplyRulesWithRemoteProgress(datasetID string, id string, aar *annotationrule.ApplyRules, onSubmitted func(string)) (*annotation.Annotation, error) {
	ann, err := a.Get(datasetID, id)
	if err != nil {
		return nil, err
	}

	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	release, err := a.acquireRuleRun(ann.DatasetID, ann.ID)
	if err != nil {
		return nil, err
	}
	defer release()

	// apply rules...
	if err := a.ruleApplier.ApplyRulesWithRemoteProgress(a.fileSysMgt.DatasetImagesDir(ds), ann, aar.Rules, onSubmitted); err != nil {
		return nil, fmt.Errorf("failed to apply annotation rules: %w", err)
	}

	ann.AppliedRules = append(ann.AppliedRules, aar.Rules...)

	if err := a.annotationStore.UpdateAnnotation(ann); err != nil {
		return nil, fmt.Errorf("failed to update annotation in store: %w", err)
	}

	return ann, nil
}

func (a *Annotation) acquireRuleRun(datasetID string, id string) (func(), error) {
	key := datasetID + "/" + id
	if _, loaded := a.activeRuleRuns.LoadOrStore(key, struct{}{}); loaded {
		return nil, ErrAnnotationRuleRunInProgress
	}
	return func() {
		a.activeRuleRuns.Delete(key)
	}, nil
}

func (a *Annotation) GetAvailableCategories(datasetID, id string) ([]string, error) {
	ann, err := a.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation: %w", err)
	}
	if !ann.Segmented {
		return nil, nil
	}
	pages, err := pagesparser.Range(ann.Pages)
	if err != nil {
		return nil, nil
	}
	if len(pages) == 0 {
		return nil, nil
	}

	categorySet := make(map[string]struct{})

	for _, page := range pages {
		af, _, altoErr := a.fileSysMgt.RetrieveAnnotationAltoPage(ann, page)
		if altoErr == nil {
			for _, ot := range af.Tags.OtherTags {
				categorySet[ot.Label] = struct{}{}
			}
		}

		md, mdErr := a.fileSysMgt.RetrieveAnnotationMarkdownPage(ann, page)
		if mdErr != nil {
			if altoErr != nil {
				return nil, fmt.Errorf("failed to retrieve markdown or ALTO annotation: [ALTO err: %w] [markdown err: %w]", altoErr, mdErr)
			}
			continue
		}
		for _, ot := range md.GetCategories() {
			categorySet[ot] = struct{}{}
		}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	return categories, nil
}

func (a *Annotation) GetAnnotationIndex(datasetID, id string, categories []string) (*annotation.Index, error) {
	ann, err := a.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation: %w", err)
	}
	pages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found for annotation %s", ann.ID)
	}
	emptyIndex := func() *annotation.Index {
		return &annotation.Index{
			DatasetID:    datasetID,
			AnnotationID: id,
			Nodes:        []*annotation.IndexNode{},
		}
	}

	var annAltoErr error
	if ann.Segmented {
		categories, allLocs, err := a.getIndexFromAnnotationAlto(pages, ann, categories)
		if err == nil {
			return &annotation.Index{
				DatasetID:    datasetID,
				AnnotationID: id,
				Nodes:        buildNodes(categories, allLocs),
			}, nil
		}
		annAltoErr = err
	} else {
		annAltoErr = fmt.Errorf("annotation %s is not segmented", ann.ID)
	}

	categories, allLocs, annMdErr := a.getIndexFromAnnotationMarkdown(pages, ann, categories)
	if annMdErr == nil {
		return &annotation.Index{
			DatasetID:    datasetID,
			AnnotationID: id,
			Nodes:        buildNodes(categories, allLocs),
		}, nil
	}

	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	if ds.EditionID == "" {
		if !ann.Segmented {
			return emptyIndex(), nil
		}
		return nil, fmt.Errorf("annotation doesn't contain ALTO or markdown files for one or more of the pages and dataset doesn't have an edition to extract index from: [annotation ALTO err: %w] [annotation markdown err: %w]", annAltoErr, annMdErr)
	}

	categories, allLocs, edAltoErr := a.getIndexFromEditionAlto(pages, ds.EditionID, categories)
	if edAltoErr == nil {
		return &annotation.Index{
			DatasetID:    datasetID,
			AnnotationID: id,
			Nodes:        buildNodes(categories, allLocs),
		}, nil
	}

	categories, allLocs, edMdErr := a.getIndexFromEditionMarkdown(pages, ds.EditionID, categories)
	if edMdErr == nil {
		return &annotation.Index{
			DatasetID:    datasetID,
			AnnotationID: id,
			Nodes:        buildNodes(categories, allLocs),
		}, nil
	}
	if !ann.Segmented && errors.Is(annMdErr, filesys.ErrMarkdownPageNotFound) && errors.Is(edMdErr, filesys.ErrMarkdownPageNotFound) {
		return emptyIndex(), nil
	}

	return nil, fmt.Errorf("failed to get annotation index, tried annotation ALTO, annotation markdown, edition ALTO, and edition markdown: [annotation ALTO err: %w] [annotation markdown err: %w] [edition ALTO err: %w] [edition markdown err: %w]", annAltoErr, annMdErr, edAltoErr, edMdErr)
}

func (a *Annotation) Delete(datasetID string, annotationID string, fileSysClean bool) error {
	ann, err := a.Get(datasetID, annotationID)
	if err != nil {
		return fmt.Errorf("failed to get annotation: %w", err)
	}
	if ann == nil {
		return fmt.Errorf("annotation not found")
	}
	if err := a.annotationStore.DeleteAnnotation(datasetID, annotationID); err != nil {
		return fmt.Errorf("failed to delete annotation from store: %w", err)
	}
	if fileSysClean {
		annoDir := a.fileSysMgt.DatasetAnnotationAltoDir(ann)
		if err := os.RemoveAll(annoDir); err != nil {
			return fmt.Errorf("failed to delete annotation files: %w", err)
		}
	}

	return nil
}

func (a *Annotation) Update(datasetID string, annotationID string, ann *annotation.Annotation) (*annotation.Annotation, error) {
	fromDB, err := a.annotationStore.GetAnnotation(datasetID, annotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation from store: %w", err)
	}
	if fromDB == nil {
		return nil, fmt.Errorf("annotation not found")
	}

	// only allow updating certain fields
	fromDB.Meta.Name = ann.Meta.Name
	fromDB.Meta.Description = ann.Meta.Description
	fromDB.GroundTruth = ann.GroundTruth
	fromDB.Hidden = ann.Hidden
	fromDB.OriginAnnotationID = ann.OriginAnnotationID
	fromDB.Pages = ann.Pages

	if err := a.annotationStore.UpdateAnnotation(fromDB); err != nil {
		return nil, fmt.Errorf("failed to update annotation in store: %w", err)
	}

	return fromDB, nil
}

func (a *Annotation) Duplicate(datasetID string, annotationID string, name string, description string, copyFeatureResults bool) (*annotation.Annotation, error) {
	origAnn, err := a.Get(datasetID, annotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original annotation: %w", err)
	}
	if origAnn == nil {
		return nil, fmt.Errorf("original annotation not found")
	}

	var ann *annotation.Annotation
	if err := deepcopy.Copy(&ann, &origAnn); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}

	ann.ID = idgen.GenerateID(store.AnnotationIDPrefix)
	ann.Meta.Name = name
	ann.Meta.Description = description

	if ann.Meta.Name == "" {
		ann.Meta.Name = "Copy of " + origAnn.Meta.Name + " " + ann.ID
	}

	if ann.Meta.Description != "" {
		ann.Meta.Description = "Copy of " + origAnn.Meta.Name + " [original description: " + origAnn.Meta.Description + "]"
	}

	if origAnn.Segmented {
		if err := futils.CopyDir(a.fileSysMgt.DatasetAnnotationAltoDir(origAnn), a.fileSysMgt.DatasetAnnotationAltoDir(ann)); err != nil {
			return nil, fmt.Errorf("failed to copy annotations for new annotation: %w", err)
		}
	}
	ann.OriginAnnotationID = origAnn.ID
	if err := a.annotationStore.InsertAnnotation(ann); err != nil {
		return nil, fmt.Errorf("failed to insert new annotation to store: %w", err)
	}
	if copyFeatureResults {
		if err := a.featureResultsSvc.CopyResults(origAnn.DatasetID, origAnn.ID, origAnn.DatasetID, ann.ID); err != nil {
			return nil, fmt.Errorf("failed to copy feature results for new annotation: %w", err)
		}
	}

	return ann, nil
}

func (a *Annotation) GetReviewByIndex(datasetID string, annotationID string, toReview *annotation.ExpectedBlocks) (*annotation.ExpectedBlocks, error) {
	ann, err := a.Get(datasetID, annotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation: %w", err)
	}

	pages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found for annotation %s", ann.ID)
	}

	var diffs []*annotation.SuggestedDiff
	for _, page := range pages {
		af, pageAltoPath, err := a.fileSysMgt.RetrieveAnnotationAltoPage(ann, fmt.Sprintf("%d", page))
		if err != nil {
			return nil, fmt.Errorf("load ALTO: %w", err)
		}

		// todo: add support for markdown...

		bp, err := alto.ExtractBlocksByCategory(af, toReview.Category)
		if err != nil {
			return nil, fmt.Errorf("failed to extract headers from ALTO page %s: %w", pageAltoPath, err)
		}
		for _, block := range bp {
			contents := alto.ExtractTextContentsFromBlock(block)
			if len(contents) == 0 {
				continue
			}
			diff := &annotation.SuggestedDiff{
				Page:        page,
				TextBlockID: block.ID,
				Old:         contents,
			}
			diffs = append(diffs, diff)
		}
	}

	if slices.Contains(toReview.SanityChecks, annotation.ExpectedBlocksSanityTypeExact) {
		if len(diffs) != len(toReview.ExpectedBlocks) {
			toReview.FailedChecks = append(toReview.FailedChecks, annotation.ExpectedBlocksSanityTypeExact)
		}
	}
	if len(toReview.FailedChecks) > 0 {
		return toReview, nil
	}

	for i, diff := range diffs {
		if i >= len(toReview.ExpectedBlocks) {
			break
		}
		diff.Correction = toReview.ExpectedBlocks[i]
	}

	for _, diff := range diffs {
		if !slices.Equal(diff.Old, diff.Correction) {
			toReview.SuggestedDiffs = append(toReview.SuggestedDiffs, diff)
		}
	}
	return toReview, nil
}

func (a *Annotation) ListAnnotationsByUsedModels() (map[string][]*annotation.Reference, error) {
	anns1, err := a.annotationStore.ListAppliedRulesByAnnotationIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations from store: %w", err)
	}
	modelToAnns := make(map[string][]*annotation.Reference)
	for datasetID, anns := range anns1 {
		for annID, appliedRules := range anns {
			modelIDs := annotationrule.ExtractModelIDsFromRules(appliedRules)
			for _, modelID := range modelIDs {
				modelToAnns[modelID] = append(modelToAnns[modelID], &annotation.Reference{
					DatasetID: datasetID,
					ID:        annID,
				})
			}
		}
	}
	return modelToAnns, nil
}

func (a *Annotation) ListAnnotationsByAnnotationReferences(refs []*annotation.Reference) ([]*annotation.Annotation, error) {
	anns, err := a.annotationStore.ListAnnotationsByAnnotationReferences(refs)
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations from store: %w", err)
	}
	if err := a.populateTranscriptionFallbacks(anns); err != nil {
		return nil, err
	}
	return anns, nil
}

func (a *Annotation) ListAnnotationsByDatasetIDs(dsIDs []string) ([]*annotation.Annotation, error) {
	anns, err := a.annotationStore.ListAnnotationsByDatasetIDs(dsIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations from store: %w", err)
	}
	if err := a.populateTranscriptionFallbacks(anns); err != nil {
		return nil, err
	}
	return anns, nil
}

func (a *Annotation) Merge(datasetID string, dstAnnID string, req annotation.MergeRequest) (*annotation.Annotation, error) {
	dstAnn, err := a.Get(datasetID, dstAnnID)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination annotation: %w", err)
	}
	toMerge, err := a.ListAnnotationsByAnnotationReferences(req.AnnotationsToMerge)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotations to merge: %w", err)
	}

	// calculate pages after merge (and verify pages do not intersect)
	var pagesInMerged []int
	dstPages, err := pagesparser.IntRange(dstAnn.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for destination annotation: %w", err)
	}
	pagesInMerged = append(pagesInMerged, dstPages...)
	dstPageSet := make(map[int]struct{})
	for _, p := range dstPages {
		dstPageSet[p] = struct{}{}
	}
	for _, ann := range toMerge {
		pages, err := pagesparser.IntRange(ann.Pages)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pages for annotation to merge: %w", err)
		}
		if intersect := lo.Intersect(dstPages, pages); len(intersect) > 0 {
			return nil, fmt.Errorf("pages %v in annotation %s overlap with destination annotation", pagesparser.ToString(intersect), ann.ID)
		}
		pagesInMerged = append(pagesInMerged, pages...)
	}

	// update fields of destination annotation based on merged annotations
	dstAnn.LinesDetected = lo.SomeBy(toMerge, func(a *annotation.Annotation) bool { return a.LinesDetected }) || dstAnn.LinesDetected
	dstAnn.Ocred = lo.SomeBy(toMerge, func(a *annotation.Annotation) bool { return a.Ocred }) || dstAnn.Ocred
	dstAnn.Segmented = lo.SomeBy(toMerge, func(a *annotation.Annotation) bool { return a.Segmented }) || dstAnn.Segmented
	dstAnn.GroundTruth = lo.EveryBy(toMerge, func(a *annotation.Annotation) bool { return a.GroundTruth }) && dstAnn.GroundTruth
	dstAnn.MergedAnnotations = append(dstAnn.MergedAnnotations, lo.Map(toMerge, func(a *annotation.Annotation, _ int) annotation.MergedReference {
		return annotation.MergedReference{
			Reference: annotation.Reference{
				DatasetID: a.DatasetID,
				ID:        a.ID,
			},
			MergedAt: time.Now(),
		}
	})...)
	dstAnn.Pages = pagesparser.ToString(pagesInMerged)

	// copy image files (if not already exist) - this is to ensure the merged annotation can be used even if the original annotations are deleted later
	for _, ann := range toMerge {
		if ann.DatasetID == datasetID {
			continue
		}
		pages, err := pagesparser.IntRange(ann.Pages)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pages for annotation to merge: %w", err)
		}
		for _, p := range pages {
			imgDstPath := path.Join(a.fileSysMgt.DatasetImagesDirByID(datasetID), pagesparser.PageToPNGFilename(p))
			if _, err := os.Stat(imgDstPath); os.IsNotExist(err) {
				imgSrcPath := path.Join(a.fileSysMgt.DatasetImagesDirByID(ann.DatasetID), pagesparser.PageToPNGFilename(p))
				if err := futils.CopyFile(imgSrcPath, imgDstPath); err != nil {
					return nil, fmt.Errorf("failed to copy image file for page %d from annotation %s: %w", p, ann.ID, err)
				}
			}
		}
	}

	// copy alto files
	for _, ann := range toMerge {
		if ann.Segmented {
			srcDir := a.fileSysMgt.DatasetAnnotationAltoDir(ann)
			if err := futils.CopyDir(srcDir, a.fileSysMgt.DatasetAnnotationAltoDir(dstAnn)); err != nil {
				return nil, fmt.Errorf("failed to copy ALTO files for merged annotation %s: %w", ann.ID, err)
			}
		}
	}

	// copy feature results
	for _, ann := range toMerge {
		if err := a.featureResultsSvc.CopyResults(ann.DatasetID, ann.ID, dstAnn.DatasetID, dstAnn.ID); err != nil {
			return nil, fmt.Errorf("failed to copy feature results for merged annotation %s: %w", ann.ID, err)
		}
	}

	// update destination annotation
	if err := a.annotationStore.UpdateAnnotation(dstAnn); err != nil {
		return nil, fmt.Errorf("failed to update destination annotation in store: %w", err)
	}

	return dstAnn, nil
}

func (a *Annotation) getIndexFromAnnotationAlto(pages []int, ann *annotation.Annotation, categories []string) ([]string, []categoryPageContent, error) {
	return getIndexFromAlto(pages, categories, func(page int) (*alto.Alto, error) {
		af, _, err := a.fileSysMgt.RetrieveAnnotationAltoPage(ann, fmt.Sprintf("%d", page))
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve annotation alto: %w", err)
		}
		return af, nil
	}, "ALTO")
}

func (a *Annotation) getIndexFromAnnotationMarkdown(pages []int, ann *annotation.Annotation, categories []string) ([]string, []categoryPageContent, error) {
	return getIndexFromMarkdown(pages, categories, func(page int) (*markdown.Markdown, error) {
		md, err := a.fileSysMgt.RetrieveAnnotationMarkdownPage(ann, fmt.Sprintf("%d", page))
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve annotation markdown page: %w", err)
		}
		return md, nil
	}, "annotation markdown")
}

func (a *Annotation) getIndexFromEditionAlto(pages []int, editionKey string, categories []string) ([]string, []categoryPageContent, error) {
	return getIndexFromAlto(pages, categories, func(page int) (*alto.Alto, error) {
		af, _, err := a.fileSysMgt.RetrieveEditionAltoPage(editionKey, page)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve edition alto: %w", err)
		}
		return af, nil
	}, "edition ALTO")
}

func (a *Annotation) getIndexFromEditionMarkdown(pages []int, editionKey string, categories []string) ([]string, []categoryPageContent, error) {
	return getIndexFromMarkdown(pages, categories, func(page int) (*markdown.Markdown, error) {
		md, err := a.fileSysMgt.RetrieveEditionMarkdownPage(editionKey, page)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve edition markdown page: %w", err)
		}
		return md, nil
	}, "edition markdown")
}

func getIndexFromAlto(pages []int, categories []string, loadPage func(int) (*alto.Alto, error), source string) ([]string, []categoryPageContent, error) {
	altoCat := categories
	if len(altoCat) == 0 {
		altoCat = []string{"MainZone-Head--Book", "MainZone-Head--Section"}
	}
	allLocs := make([]categoryPageContent, 0)
	for _, page := range pages {
		af, err := loadPage(page)
		if err != nil {
			return nil, nil, err
		}
		headers, err := alto.ExtractCategoryContents(af, altoCat, " / ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract headers from %s page %d: %w", source, page, err)
		}

		slices.SortFunc(headers, func(a, b *alto.CategoryAndContent) int {
			if a.VerticalPosition < b.VerticalPosition {
				return -1
			}
			if a.VerticalPosition > b.VerticalPosition {
				return 1
			}
			if a.HorizontalPosition < b.HorizontalPosition {
				return -1
			}
			if a.HorizontalPosition > b.HorizontalPosition {
				return 1
			}
			return 0
		})
		for _, h := range headers {
			allLocs = append(allLocs, categoryPageContent{
				page:     page,
				category: h.Category,
				content:  h.Content,
			})
		}
	}

	return altoCat, allLocs, nil
}

func getIndexFromMarkdown(pages []int, categories []string, loadPage func(int) (*markdown.Markdown, error), source string) ([]string, []categoryPageContent, error) {
	allLocs := make([]categoryPageContent, 0)
	seenCategories := make(map[string]struct{})
	for _, page := range pages {
		md, err := loadPage(page)
		if err != nil {
			return nil, nil, err
		}
		headers, err := markdown.ExtractCategoryContentsFromMarkdown(md, categories, " / ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract headers from %s page %d: %w", source, page, err)
		}
		for _, h := range headers {
			allLocs = append(allLocs, categoryPageContent{
				page:     page,
				category: h.Category,
				content:  h.Content,
			})
			seenCategories[h.Category] = struct{}{}
		}
	}
	if len(categories) == 0 {
		return orderedMarkdownHeaderCategories(seenCategories), allLocs, nil
	}
	return categories, allLocs, nil
}

func orderedMarkdownHeaderCategories(seenCategories map[string]struct{}) []string {
	categories := lo.Keys(seenCategories)
	slices.SortFunc(categories, func(a, b string) int {
		return markdownHeaderLevel(a) - markdownHeaderLevel(b)
	})
	return categories
}

func markdownHeaderLevel(category string) int {
	level, err := strconv.Atoi(strings.TrimPrefix(category, markdown.HeaderPrefix))
	if err != nil || level < 1 || level > 6 || !strings.HasPrefix(category, markdown.HeaderPrefix) {
		return 7
	}
	return level
}

func buildNodes(categories []string, data []categoryPageContent) []*annotation.IndexNode {
	categoryRank := make(map[string]int, len(categories))
	for rank, category := range categories {
		categoryRank[category] = rank
	}

	type rankedNode struct {
		rank int
		node *annotation.IndexNode
	}

	var nodes []*annotation.IndexNode
	stack := make([]rankedNode, 0, len(categories))
	for _, item := range data {
		rank, ok := categoryRank[item.category]
		if !ok {
			continue
		}

		node := &annotation.IndexNode{
			Category: item.category,
			Content:  item.content,
			Location: common.ALTOLocation{Page: fmt.Sprintf("%d", item.page)},
		}

		for len(stack) > 0 && stack[len(stack)-1].rank >= rank {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			nodes = append(nodes, node)
		} else {
			parent := stack[len(stack)-1].node
			parent.Children = append(parent.Children, node)
		}
		stack = append(stack, rankedNode{rank: rank, node: node})
	}

	return nodes
}

type categoryPageContent struct {
	page     int
	category string
	content  string
}
