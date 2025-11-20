package httpapi

import (
	_ "github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/docs" // swagger docs
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
	"net/http"
	"strings"
)

type Handlers struct {
	deps *Dependencies
}

// ListEditions godoc
// @Summary      List Editions
// @Description  Get a list of available editions. Optionally include facsimiles.
// @Tags         Editions
// @Param        expand  query     string  false  "Include related entities"  Enums(facsimiles)
// @Produce      json
// @Success      200  {array}   models.Edition
// @Router       /editions [get]
func (h *Handlers) ListEditions(r *http.Request) ([]*models.Edition, error) {
	expand := r.URL.Query().Get("expand")
	return h.deps.EditionService.ListEditions(strings.Contains(expand, "facsimiles"))
}

func NewHandlers(deps *Dependencies) *Handlers {
	return &Handlers{deps: deps}
}
