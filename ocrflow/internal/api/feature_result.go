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
// @Param keys query string false "Comma-separated list of keys to filter results"
// @Param features query string false "Comma-separated list of feature names to filter results"
// @Success 200 {array} feature.Result
// @Router  /datasets/{dataSetId}/results [get]
func (h *Handlers) ListResults(r *http.Request) (any, error) {
	dataSetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	keys := lo.Map(strings.Split(r.URL.Query().Get("keys"), ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})
	features := lo.Map(strings.Split(r.URL.Query().Get("features"), ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})

	return h.deps.FeatureResultSvc.ListResults(dataSetId, keys, features)
}

// CreateResult godoc
// @Summary Create a feature result
// @Description Create a new feature result
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param dataSetId path string true "Dataset ID"
// @Param result body feature.Result true "Feature result data"
// @Success 200 {object} feature.Result
// @Security 	 BearerAuth
// @Router  /datasets/{dataSetId}/results [post]
func (h *Handlers) CreateResult(r *http.Request) (any, error) {
	dataSetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	var result feature.Result
	if err := DecodeBody(r, &result); err != nil {
		return nil, err
	}
	result.DatasetID = dataSetId
	created, err := h.deps.FeatureResultSvc.CreateResult(&result)
	if err != nil {
		return nil, err
	}
	return created, nil
}
