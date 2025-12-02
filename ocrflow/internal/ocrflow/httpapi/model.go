package httpapi

import (
	"net/http"
)

// ListModels godoc
// @Summary      List Models
// @Description  Get a list of available models.
// @Tags         Models
// @Produce      json
// @Success      200  {array}   model.Model
// @Router       /models [get]
func (h *Handlers) ListModels(r *http.Request) (any, error) {
	return h.deps.ModelSvc.List()
}
