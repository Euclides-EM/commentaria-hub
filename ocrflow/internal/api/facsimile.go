package api

import (
	"fmt"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
)

// ListFacsimiles godoc
// @Summary      List Facsimiles (bulk get)
// @Description  Get facsimiles, optionally filtered by edition ID.
// @Tags         Facsimiles
// @Param        edition_id  query     []string  false  "Filter by edition ID"  collectionFormat(multi)
// @Produce      json
// @Success      200  {array}  model.Facsimile
// @Router       /facsimilies [get]
func (h *Handlers) ListFacsimiles(r *http.Request) (any, error) {
	editionIDs := r.URL.Query()["edition_id"]
	return h.deps.FacsimileSvc.ListFacsimiles(editionIDs)
}

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

// GetFacsimile godoc
// @Summary      Get Facsimile by ID
// @Description  Get a single facsimile by its ID.
// @Tags         Facsimiles
// @Param        id  path      string  true  "Facsimile ID"
// @Produce      json
// @Success      200  {object}  model.Facsimile
// @Failure      404  "Facsimile not found"
// @Router       /facsimilies/{id} [get]
func (h *Handlers) GetFacsimile(r *http.Request) (any, error) {
	id := r.PathValue("id")
	if id == "" {
		return nil, fmt.Errorf("missing facsimile ID")
	}
	fac, err := h.deps.FacsimileSvc.GetFacsimile(id)
	if err != nil {
		return nil, err
	}
	return fac, nil
}

// UpdateFacsimile godoc
// @Summary      Update Facsimile
// @Description  Update an existing facsimile identified by ID.
// @Tags         Facsimiles
// @Accept       json
// @Produce      json
// @Param        id          path      string  true  "Facsimile ID"
// @Param        facsimile   body      model.Facsimile  true  "Facsimile data to update"
// @Security 	 BearerAuth
// @Success      200  {object}  model.Facsimile
// @Router       /facsimilies/{id} [put]
func (h *Handlers) UpdateFacsimile(r *http.Request) (any, error) {
	id := r.PathValue("id")
	if id == "" {
		return nil, fmt.Errorf("missing facsimile ID")
	}
	var facsimile model.Facsimile
	if err := DecodeBody(r, &facsimile); err != nil {
		return nil, err
	}
	if facsimile.ID != "" && facsimile.ID != id {
		return nil, fmt.Errorf("facsimile id in body (%s) does not match id in path (%s)", facsimile.ID, id)
	}
	facsimile.ID = id
	return h.deps.FacsimileSvc.UpdateFacsimile(&facsimile)
}
