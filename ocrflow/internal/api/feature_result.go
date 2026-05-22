package api

import (
	"net/http"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
)

// ListDatasetResults godoc
// @Summary List feature results
// @Description Get a list of feature results
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param dataSetId path string true "Dataset ID"
// @Param id path string true "Annotation ID"
// @Param keys query string false "list of keys to filter results" collectionFormat(multi)
// @Param features query string false "list of feature names to filter results" collectionFormat(multi)
// @Param fallback_to_origin query bool false "Whether to fallback to results of the origin annotation if no feature results are found."
// @Success 200 {array} feature.Result
// @Router  /datasets/{dataSetId}/annotations/{id}/results [get]
func (h *Handlers) ListDatasetResults(r *http.Request) (any, error) {
	dataSetId, annotationId, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}
	keys := r.URL.Query()["keys"]
	features := r.URL.Query()["features"]
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
// @Param features query string false "list of feature IDs to filter results" collectionFormat(multi)
// @Success 200 {array} feature.Result
// @Router  /editions/{editionId}/results [get]
func (h *Handlers) ListEditionResults(r *http.Request) (any, error) {
	editionID, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}
	features := r.URL.Query()["features"]
	return h.deps.FeatureResultSvc.ListResults(feature.NewEditionExecScope(), []string{editionID}, features, false)
}

// ListFeaturesResults godoc
// @Summary List feature results
// @Description Get a list of feature results for a dataset annotation or for editions.
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param scope query string true "Feature execution scope" Enums(dataset, editions)
// @Param dataset query string false "Dataset ID, required for dataset scope"
// @Param annotation query string false "Annotation ID, required for dataset scope"
// @Param keys query string false "list of keys to filter results" collectionFormat(multi)
// @Param features query string false "list of feature names to filter results" collectionFormat(multi)
// @Param fallback_to_origin query bool false "Whether to fallback to results of the origin annotation."
// @Success 200 {array} feature.Result
// @Router  /features_results [get]
func (h *Handlers) ListFeaturesResults(r *http.Request) (any, error) {
	scope, err := extractExecScope(r)
	if err != nil {
		return nil, err
	}
	keys := r.URL.Query()["keys"]
	features := r.URL.Query()["features"]
	fallbackToOrigin, err := strconv.ParseBool(r.URL.Query().Get("fallback_to_origin"))
	if err != nil {
		fallbackToOrigin = false
	}
	return h.deps.FeatureResultSvc.ListResults(scope, keys, features, fallbackToOrigin)
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

// CreateFeaturesResult godoc
// @Summary Create feature results
// @Description Create new feature results (batch) for dataset or edition scopes.
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param result body []feature.Result true "Feature results data"
// @Param push_to_origin query bool false "Whether to push dataset results to the origin annotation."
// @Success 200 {array} feature.Result
// @Security BearerAuth
// @Router /features_results [post]
func (h *Handlers) CreateFeaturesResult(r *http.Request) (any, error) {
	user := r.Context().Value(httpwrapper.GitHubUserKey)
	userLogin := ""
	if u, ok := user.(*httpwrapper.GitHubUser); ok && u != nil {
		userLogin = u.Login
	}

	var result []*feature.Result
	if err := DecodeBody(r, &result); err != nil {
		return nil, err
	}

	for _, res := range result {
		if res == nil {
			continue
		}
		if res.Source.Resp == "" {
			res.Source = feature.ResultSource{
				Name: userLogin,
				Resp: "human",
			}
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
