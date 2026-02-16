package api

import (
	"net/http"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/samber/lo"
)

// ListResults godoc
// @Summary List feature results
// @Description Get a list of feature results
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param dataSetId path string true "Dataset ID"
// @Param id path string true "Annotation ID"
// @Param keys query string false "Comma-separated list of keys to filter results"
// @Param features query string false "Comma-separated list of feature names to filter results"
// @Success 200 {array} feature.Result
// @Router  /datasets/{dataSetId}/annotations/{id}/results [get]
func (h *Handlers) ListResults(r *http.Request) (any, error) {
	dataSetId, annotationId, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}
	keys := lo.Map(strings.Split(r.URL.Query().Get("keys"), ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})
	features := lo.Map(strings.Split(r.URL.Query().Get("features"), ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})

	return h.deps.FeatureResultSvc.ListResults(dataSetId, annotationId, keys, features)
}

// CreateResult godoc
// @Summary Create a feature result
// @Description Create a new feature result
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param dataSetId path string true "Dataset ID"
// @Param id path string true "Annotation ID"
// @Param result body feature.Result true "Feature result data"
// @Success 200 {object} feature.Result
// @Security 	 BearerAuth
// @Router  /datasets/{dataSetId}/annotations/{id}/results [post]
func (h *Handlers) CreateResult(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}
	var result feature.Result
	if err := DecodeBody(r, &result); err != nil {
		return nil, err
	}
	result.DatasetID = datasetID
	result.AnnotationID = annotationID
	created, err := h.deps.FeatureResultSvc.CreateResult(&result)
	if err != nil {
		return nil, err
	}
	return created, nil
}
