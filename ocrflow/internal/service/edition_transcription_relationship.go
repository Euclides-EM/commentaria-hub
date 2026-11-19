package service

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
	"github.com/samber/lo"
)

type EditionTranscription struct {
	editionPreferredTranscriptionStore *store.EditionPreferredAnnotationSql
	editionsSvc                        *Edition
	datasetSvc                         *Dataset
	annotationSvc                      *Annotation
}

func NewEditionTranscription(editionPreferredTranscriptionStore *store.EditionPreferredAnnotationSql, editionsSvc *Edition, datasetSvc *Dataset, annotationSvc *Annotation) *EditionTranscription {
	return &EditionTranscription{
		editionPreferredTranscriptionStore: editionPreferredTranscriptionStore,
		editionsSvc:                        editionsSvc,
		datasetSvc:                         datasetSvc,
		annotationSvc:                      annotationSvc,
	}
}

func (r *EditionTranscription) ListTranscriptionsByEditionIDs(editions []string) ([]*model.EditionTranscription, error) {
	if len(editions) == 0 {
		return nil, errors.New("no editions provided")
	}

	// check that all the edition IDs are valid
	for _, edition := range editions {
		_, err := r.editionsSvc.GetEditionByID(edition)
		if err != nil {
			return nil, err
		}
	}

	// get all datasets and group them by edition ID (only for the editions we're interested in)
	allDs, err := r.datasetSvc.List(nil, nil)
	if err != nil {
		return nil, err
	}
	dsByEditions := make(map[string][]*model.Dataset)
	for _, ds := range allDs {
		if ds.EditionID == "" {
			continue
		}
		if lo.ContainsBy(editions, func(e string) bool { return strings.EqualFold(e, ds.EditionID) }) {
			dsByEditions[ds.EditionID] = append(dsByEditions[ds.EditionID], ds)
		}
	}

	// if there are any editions that don't have any datasets, we still want to include them in the result with an empty list of annotations
	for _, edition := range editions {
		if _, ok := dsByEditions[edition]; !ok {
			dsByEditions[edition] = []*model.Dataset{}
		}
	}

	// try to get the manually configured preferred annotation for each edition (if it exists)
	preferredAnnRefsByEdition, err := r.editionPreferredTranscriptionStore.ListEditionPreferredTranscription(editions)
	if err != nil {
		return nil, err
	}

	var out []*model.EditionTranscription
	for _, edition := range editions {
		preferredAnnRef, ok := preferredAnnRefsByEdition[edition]
		preferredAnnSource := model.EditionTranscriptionPreferredAnnotationSourceManual
		if !ok {
			preferredAnnRef, err = r.getPreferredAnnotationByHeuristics(dsByEditions[edition])
			preferredAnnSource = model.EditionTranscriptionPreferredAnnotationSourceCalculated
			if err != nil {
				return nil, fmt.Errorf("failed to get preferred annotation by heuristics for edition %s: %w", edition, err)
			}
		}

		editionTranscription := &model.EditionTranscription{
			EditionID: edition,
			Datasets:  lo.Map(dsByEditions[edition], func(ds *model.Dataset, _ int) string { return ds.ID }),
		}
		if preferredAnnRef != nil {
			editionTranscription.PreferredAnnotation = &model.EditionTranscriptionPreferredAnnotation{
				Reference: *preferredAnnRef,
				Source:    preferredAnnSource,
			}
		}
		out = append(out, editionTranscription)
	}

	return out, nil
}

func (r *EditionTranscription) GetTranscriptionByEditionID(editionID string) (*model.EditionTranscription, error) {
	transcriptions, err := r.ListTranscriptionsByEditionIDs([]string{editionID})
	if err != nil {
		return nil, fmt.Errorf("failed to list transcriptions by edition ID %s: %w", editionID, err)
	}
	if len(transcriptions) == 0 {
		return nil, fmt.Errorf("no transcription found for edition ID %s", editionID)
	}
	return transcriptions[0], nil
}

func (r *EditionTranscription) getPreferredAnnotationByHeuristics(datasets []*model.Dataset) (*annotation.Reference, error) {
	if len(datasets) == 0 {
		return nil, nil
	}
	anns, err := r.annotationSvc.ListAnnotationsByDatasetIDs(lo.Map(datasets, func(ds *model.Dataset, _ int) string { return ds.ID }))
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations for datasets %v: %w", lo.Map(datasets, func(ds *model.Dataset, _ int) string { return ds.ID }), err)
	}
	if len(anns) == 0 {
		return nil, nil
	}

	// Sort logic:
	// 1. Prefer annotations that are further along in the pipeline (e.g. OCR > text line segmentation > zone segmentation > raw)
	// 2. If two annotations are at the same pipeline stage, prefer the one that covers more pages
	// 3. If two annotations are at the same pipeline stage and cover the same number of pages, prefer the one that is marked as ground truth
	slices.SortStableFunc(anns, func(a, b *annotation.Annotation) int {
		if a.PipelineStage.After(b.PipelineStage) {
			return -1
		}
		if b.PipelineStage.After(a.PipelineStage) {
			return 1
		}

		pagesInA, err := pagesparser.Range(a.Pages)
		if err != nil {
			return 0
		}
		pagesInB, err := pagesparser.Range(b.Pages)
		if err != nil {
			return 0
		}

		if len(pagesInA) > len(pagesInB) {
			return -1
		}
		if len(pagesInB) > len(pagesInA) {
			return 1
		}

		if a.GroundTruth && !b.GroundTruth {
			return -1
		}
		if b.GroundTruth && !a.GroundTruth {
			return 1
		}

		return 0
	})

	preferredAnn := anns[len(anns)-1]
	return &annotation.Reference{DatasetID: preferredAnn.DatasetID, ID: preferredAnn.ID}, nil
}

func (r *EditionTranscription) Update(editionID string, et *model.EditionTranscription) (*model.EditionTranscription, error) {
	existing, err := r.GetTranscriptionByEditionID(editionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing transcription for edition ID %s: %w", editionID, err)
	}

	if et.PreferredAnnotation == nil || et.PreferredAnnotation.DatasetID == "" || et.PreferredAnnotation.ID == "" {
		return existing, nil
	}

	if !lo.ContainsBy(existing.Datasets, func(dsID string) bool { return dsID == et.PreferredAnnotation.DatasetID }) {
		return nil, fmt.Errorf("preferred annotation dataset ID %s is not associated with edition ID %s", et.PreferredAnnotation.DatasetID, editionID)
	}

	ann, err := r.annotationSvc.Get(et.PreferredAnnotation.DatasetID, et.PreferredAnnotation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation with dataset ID %s and annotation ID %s: %w", et.PreferredAnnotation.DatasetID, et.PreferredAnnotation.ID, err)
	}

	existing.PreferredAnnotation = &model.EditionTranscriptionPreferredAnnotation{
		Reference: annotation.Reference{
			DatasetID: ann.DatasetID,
			ID:        ann.ID,
		},
		Source: model.EditionTranscriptionPreferredAnnotationSourceManual,
	}

	err = r.editionPreferredTranscriptionStore.UpsertEditionPreferredTranscription(editionID, &existing.PreferredAnnotation.Reference)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert preferred annotation for edition ID %s: %w", editionID, err)
	}

	return existing, nil
}
