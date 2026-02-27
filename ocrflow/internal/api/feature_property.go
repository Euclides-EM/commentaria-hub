package api

import "net/http"

// ListFeatureProperties godoc
// @Summary List available feature properties
// @Description Get a list of all feature property keys
// @Tags Feature
// @Produce json
// @Success 200 {array} string
// @Router /features/properties [get]
func (h *Handlers) ListFeatureProperties(r *http.Request) (any, error) {
	return h.deps.FeaturePropertySvc.ListFeaturePropertyKeys()
}
