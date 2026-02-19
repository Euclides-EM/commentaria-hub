package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
)

// CreateAnnotationReview godoc
// @Summary Create an annotation review based on expected blocks
// @Description Create an annotation review by providing expected blocks for comparison
// @Tags Annotations
// @Accept json
// @Produce json
// @Param dataSetId path string true "Dataset ID"
// @Param id path string true "Annotation ID"
// @Param review body annotation.ExpectedBlocks true "Expected blocks for review"
// @Security 	 BearerAuth
// @Success 200 {object} annotation.ExpectedBlocks
// @Router /datasets/{dataSetId}/annotations/{id}/review [post]
func (h *Handlers) CreateAnnotationReview(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	var toReview *annotation.ExpectedBlocks
	if err = DecodeBody(r, &toReview); err != nil {
		return nil, err
	}

	reviewData, err := h.deps.AnnotationSvc.GetReviewByIndex(datasetID, annotationID, toReview)
	if err != nil {
		return nil, err
	}

	return reviewData, nil
}
