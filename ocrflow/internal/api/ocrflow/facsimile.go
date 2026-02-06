package ocrflow

import (
	"errors"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
)

// CreateFacsimile godoc
// @Summary      Create Facsimile
// @Description  Create a new facsimile
// @Tags         Facsimiles
// @Accept       json
// @Produce      json
// @Param        editionId  path      string          true  "Edition ID"
// @Param        facsimile  body      ocrflow.Facsimile  true  "Facsimile to create"
// @Security 	 BearerAuth
// @Success      200  {object}  ocrflow.Facsimile
// @Router       /editions/{editionId}/facsimilies [post]
func (h *Handlers) CreateFacsimile(r *http.Request) (any, error) {
	editionId := r.PathValue("editionId")
	if editionId == "" {
		return nil, errors.New("missing edition ID")
	}
	var facsimile ocrflow.Facsimile
	if err := common.DecodeBody(r, &facsimile); err != nil {
		return nil, err
	}
	return h.deps.EditionSvc.CreateFacsimile(editionId, &facsimile)
}
