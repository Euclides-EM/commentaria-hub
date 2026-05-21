package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	"github.com/samber/lo"
)

// ListDatasetResults godoc
// @Summary List feature results
// @Description Get a list of feature results
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param dataSetId path string true "Dataset ID"
// @Param id path string true "Annotation ID"
// @Param keys query string false "Comma-separated list of keys to filter results"
// @Param features query string false "Comma-separated list of feature names to filter results"
// @Param fallback_to_origin query bool false "Whether to fallback to results of the origin annotation if no feature results are found."
// @Success 200 {array} feature.Result
// @Router  /datasets/{dataSetId}/annotations/{id}/results [get]
func (h *Handlers) ListDatasetResults(r *http.Request) (any, error) {
	dataSetId, annotationId, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}
	var keys, features []string
	if keysStr := r.URL.Query().Get("keys"); keysStr != "" {
		keys = lo.Map(strings.Split(keysStr, ","), func(s string, _ int) string {
			return strings.TrimSpace(s)
		})
	}
	if featuresStr := r.URL.Query().Get("features"); featuresStr != "" {
		features = lo.Map(strings.Split(featuresStr, ","), func(s string, _ int) string {
			return strings.TrimSpace(s)
		})
	}
	fallbackToOrigin, err := strconv.ParseBool(r.URL.Query().Get("fallback_to_origin"))
	if err != nil {
		fallbackToOrigin = false
	}
	return h.deps.FeatureResultSvc.ListResults(feature.NewDatasetExecScope(dataSetId, annotationId), keys, features, fallbackToOrigin)
}

// ListEditionResults godoc
// @Summary List edition feature results
// @Description Get a list of feature results for an edition
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param editionId path string true "Edition ID"
// @Param features query string false "Comma-separated list of feature IDs to filter results"
// @Success 200 {array} feature.Result
// @Router  /editions/{editionId}/results [get]
func (h *Handlers) ListEditionResults(r *http.Request) (any, error) {
	editionID, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}
	var features []string
	if featuresStr := r.URL.Query().Get("features"); featuresStr != "" {
		features = lo.Map(strings.Split(featuresStr, ","), func(s string, _ int) string {
			return strings.TrimSpace(s)
		})
	}
	return h.deps.FeatureResultSvc.ListResults(feature.NewEditionExecScope(), []string{editionID}, features, false)
}

// CreateDatasetResult godoc
// @Summary Create feature results
// @Description Create new feature results (batch)
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param dataSetId path string true "Dataset ID"
// @Param id path string true "Annotation ID"
// @Param result body []feature.Result true "Feature results data"
// @Param push_to_origin query bool false "Whether to push the created results to the origin annotation."
// @Success 200 {array} feature.Result
// @Security BearerAuth
// @Router /datasets/{dataSetId}/annotations/{id}/results [post]
func (h *Handlers) CreateDatasetResult(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	user := r.Context().Value(httpwrapper.GitHubUserKey)
	userLogin := ""
	if u, ok := user.(*httpwrapper.GitHubUser); ok && u != nil {
		userLogin = u.Login
	}

	var result []*feature.Result
	// Pass the pointer to the slice (&result)
	if err := DecodeBody(r, &result); err != nil {
		return nil, err
	}

	for _, res := range result {
		res.Scope.DatasetID = datasetID
		res.Scope.AnnotationID = annotationID
		res.Source = feature.ResultSource{
			Name: userLogin,
			Resp: "human",
		}
	}

	pushToOrigin, err := strconv.ParseBool(r.URL.Query().Get("push_to_origin"))
	if err != nil {
		pushToOrigin = false
	}

	created, err := h.deps.FeatureResultSvc.CreateResult(result, pushToOrigin)
	if err != nil {
		return nil, err
	}
	return created, nil
}
