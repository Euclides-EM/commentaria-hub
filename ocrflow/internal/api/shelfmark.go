package api

import (
	"fmt"
	"net/http"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
)

// ListShelfmarks godoc
// @Summary      List shelfmarks
// @Description  Lists shelfmarks, optionally filtered by edition ID.
// @Tags         Shelfmarks
// @Produce      json
// @Param        edition_id  query     []string  false  "Filter by edition ID"  collectionFormat(multi)
// @Success      200  {array}  model.EditionShelfmark
// @Router       /shelfmarks [get]
func (h *Handlers) ListAllShelfmarks(r *http.Request) (any, error) {
	editionIDs := r.URL.Query()["edition_id"]
	if len(editionIDs) > 0 {
		return h.deps.ShelfmarkSvc.ListShelfmarksByEditionIDs(editionIDs)
	}
	return h.deps.ShelfmarkSvc.ListAllShelfmarks()
}

// ListShelfmarks godoc
// @Summary      List edition shelfmarks
// @Description  Lists shelfmarks for an edition.
// @Tags         Shelfmarks
// @Produce      json
// @Param        editionId  path      string  true  "Edition ID"
// @Success      200  {array}  model.EditionShelfmark
// @Router       /editions/{editionId}/shelfmarks [get]
func (h *Handlers) ListShelfmarks(r *http.Request) (any, error) {
	editionID, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}
	return h.deps.ShelfmarkSvc.ListShelfmarks(editionID)
}

// UpsertShelfmark godoc
// @Summary      Upsert edition shelfmark
// @Description  Creates or updates a shelfmark for an edition.
// @Tags         Shelfmarks
// @Accept       json
// @Produce      json
// @Param        editionId  path      string  true  "Edition ID"
// @Param        shelfmark  body      model.EditionShelfmark  true  "Shelfmark"
// @Security 	 BearerAuth
// @Success      200  {object}  model.EditionShelfmark
// @Router       /editions/{editionId}/shelfmarks [post]
func (h *Handlers) UpsertShelfmark(r *http.Request) (any, error) {
	editionID, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}
	var shelfmark model.EditionShelfmark
	if err := DecodeBody(r, &shelfmark); err != nil {
		return nil, err
	}
	return h.deps.ShelfmarkSvc.UpsertShelfmark(editionID, &shelfmark)
}

// UpdateShelfmark godoc
// @Summary      Update edition shelfmark
// @Description  Updates a shelfmark for an edition.
// @Tags         Shelfmarks
// @Accept       json
// @Produce      json
// @Param        editionId    path      string  true  "Edition ID"
// @Param        shelfmarkId  path      string  true  "Shelfmark ID"
// @Param        shelfmark    body      model.EditionShelfmark  true  "Shelfmark"
// @Security 	 BearerAuth
// @Success      200  {object}  model.EditionShelfmark
// @Router       /editions/{editionId}/shelfmarks/{shelfmarkId} [put]
func (h *Handlers) UpdateShelfmark(r *http.Request) (any, error) {
	editionID, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}
	shelfmarkID := r.PathValue("shelfmarkId")
	if shelfmarkID == "" {
		return nil, fmt.Errorf("missing shelfmark ID")
	}
	var shelfmark model.EditionShelfmark
	if err := DecodeBody(r, &shelfmark); err != nil {
		return nil, err
	}
	if shelfmark.ID != "" && shelfmark.ID != shelfmarkID {
		return nil, fmt.Errorf("shelfmark id in body (%s) does not match id in path (%s)", shelfmark.ID, shelfmarkID)
	}
	shelfmark.ID = shelfmarkID
	return h.deps.ShelfmarkSvc.UpsertShelfmark(editionID, &shelfmark)
}

// DeleteShelfmark godoc
// @Summary      Delete edition shelfmark
// @Description  Deletes a shelfmark from an edition.
// @Tags         Shelfmarks
// @Param        editionId    path      string  true  "Edition ID"
// @Param        shelfmarkId  path      string  true  "Shelfmark ID"
// @Security 	 BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /editions/{editionId}/shelfmarks/{shelfmarkId} [delete]
func (h *Handlers) DeleteShelfmark(r *http.Request) (any, error) {
	editionID, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}
	shelfmarkID := r.PathValue("shelfmarkId")
	if shelfmarkID == "" {
		return nil, fmt.Errorf("missing shelfmark ID")
	}
	if err := h.deps.ShelfmarkSvc.DeleteShelfmark(editionID, shelfmarkID); err != nil {
		return nil, err
	}
	return map[string]string{"message": "shelfmark deleted"}, nil
}
