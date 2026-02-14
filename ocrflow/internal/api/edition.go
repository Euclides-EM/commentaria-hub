package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
)

// ListEditions godoc
// @Summary      List Editions
// @Description  Get a list of available editions. Optionally include facsimiles.
// @Tags         Editions
// @Param expand query []string false "Include related entities" Enums(facsimiles) collectionFormat(multi)
// @Param        orderBy query     string  false  "Order by field"            Enums(suggested)
// @Produce      json
// @Success      200  {array}   ocrflow.Edition
// @Router       /editions [get]
func (h *Handlers) ListEditions(r *http.Request) (any, error) {
	return h.deps.EditionSvc.ListEditions(model.ToEditionExpandOptions(r.URL.Query().Get("expand")), model.ToEditionOrderByOptions("orderBy"))
}

// CreateEdition godoc
// @Summary      Create Edition
// @Description  Create a new edition
// @Tags         Editions
// @Accept       json
// @Produce      json
// @Param        edition  body      ocrflow.Edition  true  "Edition to create"
// @Security 	 BearerAuth
// @Success      200  {object}  ocrflow.Edition
// @Router       /editions [post]
func (h *Handlers) CreateEdition(r *http.Request) (any, error) {
	var ed model.Edition
	if err := DecodeBody(r, &ed); err != nil {
		return nil, err
	}
	if err := h.deps.EditionSvc.CreateEdition(&ed); err != nil {
		return nil, err
	}
	return ed, nil
}
