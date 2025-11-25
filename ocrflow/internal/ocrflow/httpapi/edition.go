package httpapi

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"net/http"
)

// ListEditions godoc
// @Summary      List Editions
// @Description  Get a list of available editions. Optionally include facsimiles.
// @Tags         Editions
// @Param        expand  query     string  false  "Include related entities"  Enums(facsimiles)
// @Param        orderBy query     string  false  "Order by field"            Enums(suggested)
// @Produce      json
// @Success      200  {array}   model.Edition
// @Router       /editions [get]
func (h *Handlers) ListEditions(r *http.Request) (any, error) {
	return h.deps.EditionSvc.ListEditions(model.ToEditionExpandOptions(r.URL.Query().Get("expand")), model.ToEditionOrderByOptions("orderBy"))
}
