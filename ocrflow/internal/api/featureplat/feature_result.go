package featureplat

import (
	"net/http"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	"github.com/samber/lo"
)

// ListResults godoc
// @Summary List feature results
// @Description Get a list of feature results
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param keys query string false "Comma-separated list of keys to filter results"
// @Param features query string false "Comma-separated list of feature names to filter results"
// @Success 200 {array} featureplat.FeatureResult
// @Router /results [get]
func (h *Handlers) ListResults(r *http.Request) (any, error) {
	keys := lo.Map(strings.Split(r.URL.Query().Get("keys"), ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})
	features := lo.Map(strings.Split(r.URL.Query().Get("features"), ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})

	return h.deps.FeatureResultSvc.ListResults(keys, features)
}

// CreateResult godoc
// @Summary Create a feature result
// @Description Create a new feature result
// @Tags Feature Results
// @Accept json
// @Produce json
// @Param result body featureplat.FeatureResult true "Feature result data"
// @Success 200 {object} featureplat.FeatureResult
// @Router /results [post]
func (h *Handlers) CreateResult(r *http.Request) (any, error) {
	var result featureplat.FeatureResult
	if err := common.DecodeBody(r, &result); err != nil {
		return nil, err
	}
	created, err := h.deps.FeatureResultSvc.CreateResult(&result)
	if err != nil {
		return nil, err
	}
	return created, nil
}
