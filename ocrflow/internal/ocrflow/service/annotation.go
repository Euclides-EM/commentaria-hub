package service

import (
	"fmt"
	"os"
	"path"
	"slices"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/tiendc/go-deepcopy"
)

// todo: add interfaces to all services

type Annotation struct {
	datasetSvc      *Dataset
	ruleApplier     *AnnotationRuleApplier
	fileSysMgt      *filesys.Manager
	annotationStore *store.AnnotationSQL
}

func NewAnnotationsService(datasetSvc *Dataset, ruleApplier *AnnotationRuleApplier, fileSysMgt *filesys.Manager, annotationStore *store.AnnotationSQL) *Annotation {
	return &Annotation{
		datasetSvc:      datasetSvc,
		ruleApplier:     ruleApplier,
		fileSysMgt:      fileSysMgt,
		annotationStore: annotationStore,
	}
}

func (a *Annotation) ListAnnotations(id string) ([]*model.Annotation, error) {
	annotations, err := a.annotationStore.ListAnnotationsByDatasetID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations from store: %w", err)
	}
	return annotations, nil
}

func (a *Annotation) Get(datasetId, id string) (*model.Annotation, error) {
	annotation, err := a.annotationStore.GetAnnotation(datasetId, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation from store: %w", err)
	}
	if annotation == nil || annotation.DatasetID != datasetId {
		return nil, fmt.Errorf("annotation not found")
	}
	return annotation, nil
}

func (a *Annotation) Create(datasetID string, ann *model.Annotation) (*model.Annotation, error) {
	// validate dataset exists
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	imgPath := a.fileSysMgt.DatasetImagesDir(ds)
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no images found for dataset %s", datasetID)
	}

	// assign basic fields
	ann.ID = idgen.GenerateID(store.AnnotationIDPrefix)
	ann.DatasetID = datasetID

	if ann.Pages == "" {
		// infer pages from existing images
		pages, err := store.InferPages(imgPath, model.AnnotationFormatAlto)
		if err != nil {
			return nil, fmt.Errorf("failed to infer pages from existing dataset images: %w", err)
		}
		ann.Pages = pagesparser.ToString(pages)
	}

	// verify page images exist for all specified pages
	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages: %w", err)
	}

	for _, p := range pages {
		filename := pagesparser.PageToPNGFilename(p)
		if _, err := os.Stat(path.Join(imgPath, filename)); err != nil {
			return nil, fmt.Errorf("no such file %s in existing dataset", filename)
		}
	}
	if err := a.annotationStore.InsertAnnotation(ann); err != nil {
		return nil, fmt.Errorf("failed to insert annotation to store: %w", err)
	}
	return ann, nil
}

func (a *Annotation) CreateFromZip(aum *model.AnnotationUploadMetadata, save func(dstPath string) error) (*model.Annotation, error) {
	_, err := a.datasetSvc.Get(aum.DatasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	ann := &model.Annotation{
		Meta:               model.NewMeta(idgen.GenerateID(store.AnnotationIDPrefix)).WithName(aum.Name).WithDescription(aum.Description),
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
	dstPath := a.fileSysMgt.DatasetAnnotationAltoDir(ann)
	if aum.Format == model.AnnotationFormatYolo {
		dstPath = a.fileSysMgt.DatasetAnnotationYoloDir(ann)
	}
	if err := save(dstPath); err != nil {
		return nil, fmt.Errorf("failed to store uploaded annotations: %w", err)
	}
	if aum.Format == model.AnnotationFormatYolo {
		if err := formatcov.Yolo2Alto(a.fileSysMgt.DatasetAnnotationYoloDir(ann), a.fileSysMgt.DatasetAnnotationAltoDir(ann)); err != nil {
			return nil, fmt.Errorf("failed to convert YOLO annotations to ALTO: %w", err)
		}
	}

	pages, err := store.InferPages(a.fileSysMgt.DatasetAnnotationAltoDir(ann), model.AnnotationFormatAlto)
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

func (a *Annotation) ApplyRules(datasetID string, id string, aar *annotationrule.ApplyRules) (*model.Annotation, error) {
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
		ann, err = a.Duplicate(datasetID, id, "", "")
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
		return nil, fmt.Errorf("no ALTO directory found for annotation %s", ann.ID)
	}
	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found for annotation %s", ann.ID)
	}

	categorySet := make(map[string]struct{})

	for _, page := range pages {
		af, _, err := a.fileSysMgt.RetrieveAltoPage(ann, page)
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

func (a *Annotation) GetAnnotationIndex(datasetID, id string, categories []string) (*model.AnnotationIndex, error) {
	ann, err := a.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation: %w", err)
	}
	if !ann.Segmented {
		return nil, fmt.Errorf("no ALTO directory found for annotation %s", ann.ID)
	}
	pages, err := pagesparser.Parse(ann.Pages)
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
		af, _, err := a.fileSysMgt.RetrieveAltoPage(ann, page)
		if err != nil {
			return nil, fmt.Errorf("load ALTO: %w", err)
		}

		headers, err := alto.ExtractCategoryContents(af, categories, " / ")
		if err != nil {
			return nil, fmt.Errorf("failed to extract headers from ALTO page %d: %w", page, err)
		}

		for _, h := range headers {
			allLocs = append(allLocs, categoryPageContent{
				page:     page,
				category: h.Category,
				content:  h.Content,
			})
		}
	}

	return &model.AnnotationIndex{
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

func (a *Annotation) Update(datasetID string, annotationID string, ann *model.Annotation) (*model.Annotation, error) {
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

	if err := a.annotationStore.UpdateAnnotation(fromDB); err != nil {
		return nil, fmt.Errorf("failed to update annotation in store: %w", err)
	}

	return fromDB, nil
}

func (a *Annotation) Duplicate(datasetID string, annotationID string, name string, description string) (*model.Annotation, error) {
	origAnn, err := a.Get(datasetID, annotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original annotation: %w", err)
	}
	if origAnn == nil {
		return nil, fmt.Errorf("original annotation not found")
	}

	var ann *model.Annotation
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

func (a *Annotation) GetReviewByIndex(datasetID string, annotationID string, toReview *model.AnnotationExpectedBlocks) (*model.AnnotationExpectedBlocks, error) {
	ann, err := a.Get(datasetID, annotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation: %w", err)
	}

	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for annotation %s: %w", ann.ID, err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found for annotation %s", ann.ID)
	}

	var diffs []*model.SuggestedDiff
	for _, page := range pages {
		af, pageAltoPath, err := a.fileSysMgt.RetrieveAltoPage(ann, page)
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
			diff := &model.SuggestedDiff{
				Page:        page,
				TextBlockID: block.ID,
				Old:         contents,
			}
			diffs = append(diffs, diff)
		}
	}

	if slices.Contains(toReview.SanityChecks, model.AnnotationExpectedBlocksSanityTypeExact) {
		if len(diffs) != len(toReview.ExpectedBlocks) {
			toReview.FailedChecks = append(toReview.FailedChecks, model.AnnotationExpectedBlocksSanityTypeExact)
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

func buildNodes(remainingCats []string, data []categoryPageContent) []*model.AnnotationIndexNode {
	// Base case: if no more categories to nest or no data left
	if len(remainingCats) == 0 || len(data) == 0 {
		return nil
	}

	currentCat := remainingCats[0]
	nextCats := remainingCats[1:]

	var nodes []*model.AnnotationIndexNode

	var wipNode *model.AnnotationIndexNode
	var wipNodeFirstChildIndex int

	for i, item := range data {
		if item.category == currentCat {
			if wipNode != nil {
				if wipNodeFirstChildIndex != i {
					wipNode.Children = buildNodes(nextCats, data[wipNodeFirstChildIndex:i])
				}
				nodes = append(nodes, wipNode)
			}
			wipNode = &model.AnnotationIndexNode{
				Category: item.category,
				Content:  item.content,
				Location: model.AnnotationLocation{Page: item.page},
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
