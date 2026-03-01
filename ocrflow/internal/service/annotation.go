package service

import (
	"fmt"
	"os"
	"path"
	"slices"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/name"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/samber/lo"
	"github.com/tiendc/go-deepcopy"
)

// todo: add interfaces to all services

type Annotation struct {
	datasetSvc      *Dataset
	datasetImgSvc   *DatasetImg
	ruleApplier     *AnnotationRuleApplier
	fileSysMgt      *filesys.Manager
	annotationStore *store.AnnotationSQL
}

func NewAnnotationsService(datasetSvc *Dataset, datasetImgSvc *DatasetImg, ruleApplier *AnnotationRuleApplier, fileSysMgt *filesys.Manager, annotationStore *store.AnnotationSQL) *Annotation {
	return &Annotation{
		datasetSvc:      datasetSvc,
		datasetImgSvc:   datasetImgSvc,
		ruleApplier:     ruleApplier,
		fileSysMgt:      fileSysMgt,
		annotationStore: annotationStore,
	}
}

func (a *Annotation) ListAnnotations(id string) ([]*annotation.Annotation, error) {
	annotations, err := a.annotationStore.ListAnnotationsByDatasetID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations from store: %w", err)
	}
	return annotations, nil
}

func (a *Annotation) Get(datasetId, id string) (*annotation.Annotation, error) {
	annotation, err := a.annotationStore.GetAnnotation(datasetId, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation from store: %w", err)
	}
	if annotation == nil || annotation.DatasetID != datasetId {
		return nil, fmt.Errorf("annotation not found")
	}
	return annotation, nil
}

func (a *Annotation) Create(datasetID string, ann *annotation.Annotation) (*annotation.Annotation, error) {
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
		if datasetID == "tps" {
			if _, err := a.datasetImgSvc.GetImageMetadata(datasetID, p); err != nil {
				return nil, fmt.Errorf("failed to get image metadata for page %s: %w", p, err)
			}
		} else {
			filename := pagesparser.PageOrKeyToPNGFilename(p)
			if _, err := os.Stat(path.Join(imgPath, filename)); err != nil {
				return nil, fmt.Errorf("no such file %s in existing dataset", filename)
			}
		}
	}
	if err := a.annotationStore.InsertAnnotation(ann); err != nil {
		return nil, fmt.Errorf("failed to insert annotation to store: %w", err)
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
		ann.AppliedRules = append(ann.AppliedRules, annotationrule.NewSegment(aum.SegmentModelID))
	}
	if aum.Ocred && aum.OCRModelID != "" {
		ann.AppliedRules = append(ann.AppliedRules, annotationrule.NewDetectText(aum.OCRModelID))
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

func (a *Annotation) ApplyRules(datasetID string, id string, aar *annotationrule.ApplyRules) (*annotation.Annotation, error) {
	ann, err := a.annotationStore.GetAnnotation(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation from store: %w", err)
	}
	if ann == nil {
		return nil, fmt.Errorf("annotation not found")
	}
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	if aar.Action == annotationrule.ApplyRulesActionCreateNew {
		ann, err = a.Duplicate(datasetID, id, aar.Name, aar.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to duplicate annotation for applying rules: %w", err)
		}
	}

	// apply rules...
	if err := a.ruleApplier.ApplyRules(a.fileSysMgt.DatasetImagesDir(ds), ann, aar.Rules); err != nil {
		return nil, fmt.Errorf("failed to apply annotation rules: %w", err)
	}

	ann.AppliedRules = append(ann.AppliedRules, aar.Rules...)

	if err := a.annotationStore.UpdateAnnotation(ann); err != nil {
		return nil, fmt.Errorf("failed to update annotation in store: %w", err)
	}

	return ann, nil
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
		af, _, err := a.fileSysMgt.RetrieveAnnotationAltoPage(ann, page)
		if err != nil {
			return nil, err
		}

		for _, ot := range af.Tags.OtherTags {
			categorySet[ot.Label] = struct{}{}
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
	if !ann.Segmented {
		return nil, fmt.Errorf("no ALTO directory found for annotation %s", ann.ID)
	}
	pages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found for annotation %s", ann.ID)
	}
	if len(categories) == 0 {
		categories = []string{"MainZone-Head--Book", "MainZone-Head--Section"}
	}

	allLocs := make([]categoryPageContent, 0)

	for _, page := range pages {
		af, _, err := a.fileSysMgt.RetrieveAnnotationAltoPage(ann, fmt.Sprintf("%d", page))
		if err != nil {
			return nil, fmt.Errorf("load ALTO: %w", err)
		}

		headers, err := alto.ExtractCategoryContents(af, categories, " / ")
		if err != nil {
			return nil, fmt.Errorf("failed to extract headers from ALTO page %d: %w", page, err)
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

	return &annotation.Index{
		DatasetID:    datasetID,
		AnnotationID: id,
		Nodes:        buildNodes(categories, allLocs),
	}, nil
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
	fromDB.OriginAnnotationID = ann.OriginAnnotationID

	if err := a.annotationStore.UpdateAnnotation(fromDB); err != nil {
		return nil, fmt.Errorf("failed to update annotation in store: %w", err)
	}

	return fromDB, nil
}

func (a *Annotation) Duplicate(datasetID string, annotationID string, name string, description string) (*annotation.Annotation, error) {
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

func (a *Annotation) ListAnnotationIDsByUsedModels() (map[string][]*model.AnnotationReference, error) {
	anns1, err := a.annotationStore.ListAppliedRulesByAnnotationIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations from store: %w", err)
	}
	modelToAnns := make(map[string][]*model.AnnotationReference)
	for datasetID, anns := range anns1 {
		for annID, appliedRules := range anns {
			modelIDs := annotationrule.ExtractModelIDsFromRules(appliedRules)
			for _, modelID := range modelIDs {
				modelToAnns[modelID] = append(modelToAnns[modelID], &model.AnnotationReference{
					DatasetID: datasetID,
					ID:        annID,
				})
			}
		}
	}
	return modelToAnns, nil
}

func buildNodes(remainingCats []string, data []categoryPageContent) []*annotation.IndexNode {
	// Base case: if no more categories to nest or no data left
	if len(remainingCats) == 0 || len(data) == 0 {
		return nil
	}

	currentCat := remainingCats[0]
	nextCats := remainingCats[1:]

	var nodes []*annotation.IndexNode

	var wipNode *annotation.IndexNode
	var wipNodeFirstChildIndex int

	for i, item := range data {
		if item.category == currentCat {
			if wipNode != nil {
				if wipNodeFirstChildIndex != i {
					wipNode.Children = buildNodes(nextCats, data[wipNodeFirstChildIndex:i])
				}
				nodes = append(nodes, wipNode)
			}
			wipNode = &annotation.IndexNode{
				Category: item.category,
				Content:  item.content,
				Location: common.ALTOLocation{Page: item.page},
			}
			wipNodeFirstChildIndex = i + 1
		}
	}
	// Handle the last wipNode if exists
	if wipNode != nil {
		if wipNodeFirstChildIndex != len(data) {
			wipNode.Children = buildNodes(nextCats, data[wipNodeFirstChildIndex:])
		}
		nodes = append(nodes, wipNode)
	}

	return nodes
}

type categoryPageContent struct {
	page     int
	category string
	content  string
}
