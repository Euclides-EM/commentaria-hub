package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
)

// CreateFacsimile godoc
// @Summary      Create Facsimile
// @Description  Create a new facsimile
// @Tags         Facsimiles
// @Accept       json
// @Produce      json
// @Param        facsimile  body      model.Facsimile  true  "Facsimile to create"
// @Security 	 BearerAuth
// @Success      200  {object}  model.Facsimile
// @Router       /facsimilies [post]
func (h *Handlers) CreateFacsimile(r *http.Request) (any, error) {
	var facsimile model.Facsimile
	if err := DecodeBody(r, &facsimile); err != nil {
		return nil, err
	}
	return h.deps.FacsimileSvc.CreateFacsimile(&facsimile)
}
