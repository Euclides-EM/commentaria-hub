package httpapi

import "net/http"

// ListSuggestedRulesForDataset godoc
// @Summary      List Suggested Annotation Rules for Dataset
// @Description  Get a list of suggested annotation rules for a specific dataset.
// @Tags         Datasets
// @Param        dataSetId  path      string  true  "Dataset ID"
// @Produce      json
// @Success      200  {array}   []annotationrule.AnnotationRule
// @Router       /datasets/{dataSetId}/suggested_rules [get]
func (h *Handlers) ListSuggestedRulesForDataset(r *http.Request) (any, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}

	return h.deps.DatasetSvc.ListSuggestedAnnotationRules(datasetID)
}

// ListSuggestedReviewForDataset godoc
// @Summary      List Suggested Annotation Reviews for Dataset
// @Description  Get a list of suggested annotation reviews for a specific dataset.
// @Tags         Datasets
// @Param        dataSetId  path      string  true  "Dataset ID"
// @Produce      json
// @Success      200  {array}   []model.AnnotationExpectedBlocks
// @Router       /datasets/{dataSetId}/suggested_reviews [get]
func (h *Handlers) ListSuggestedReviewForDataset(r *http.Request) (any, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}

	return h.deps.DatasetSvc.ListSuggestedAnnotationReview(datasetID)
}
