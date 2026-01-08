package httpapi

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
)

// CreateAnnotationReview godoc
// @Summary Create an annotation review based on expected blocks
// @Description Create an annotation review by providing expected blocks for comparison
// @Tags Annotations
// @Accept json
// @Produce json
// @Param dataSetId path string true "Dataset ID"
// @Param id path string true "Annotation ID"
// @Param review body model.AnnotationExpectedBlocks true "Expected blocks for review"
// @Security 	 BearerAuth
// @Success 200 {object} model.AnnotationExpectedBlocks
// @Router /datasets/{dataSetId}/annotations/{id}/review [post]
func (h *Handlers) CreateAnnotationReview(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	var toReview *model.AnnotationExpectedBlocks
	if err = decodeBody(r, &toReview); err != nil {
		return nil, err
	}

	reviewData, err := h.deps.AnnotationSvc.GetReviewByIndex(datasetID, annotationID, toReview)
	if err != nil {
		return nil, err
	}

	return reviewData, nil
}
