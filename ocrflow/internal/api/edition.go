package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/search"
)

const (
	defaultListLimit = 20
	maxListLimit     = 5000
)

// ListEditions godoc
// @Summary      List Editions
// @Description  Get a paginated list of editions. Filter by corpus; use offset/limit for paging.
// @Tags         Editions
// @Produce      json
// @Accept       json
// @Param        edition  body  search.Query  false  "Filter, ordering, and pagination options"
// @Success      200  {object}  model.EditionListResult
// @Router       /editions/search [post]
func (h *Handlers) ListEditions(r *http.Request) (any, error) {
	var query search.Query
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return nil, fmt.Errorf("failed to decode request body: %w", err)
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	limit := query.Limit
	if limit <= 0 {
		limit = defaultListLimit
	} else if limit > maxListLimit {
		limit = maxListLimit
	}
	return h.deps.EditionSvc.ListEditions(query.FilterFunc(), query.OrderByFunc(), offset, limit)
}

// GetEdition godoc
// @Summary      Get Edition by ID
// @Description  Get a single edition by its ID.
// @Tags         Editions
// @Param        editionId  path      string  true  "Edition ID"
// @Produce      json
// @Success      200  {object}  model.Edition
// @Failure      404  "Edition not found"
// @Router       /editions/{editionId} [get]
func (h *Handlers) GetEdition(r *http.Request) (any, error) {
	key, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}
	ed, err := h.deps.EditionSvc.GetEditionByID(key)
	if err != nil {
		return nil, err
	}
	if ed == nil {
		return nil, fmt.Errorf("edition not found: %s", key)
	}
	return ed, nil
}

// CreateEdition godoc
// @Summary      Create Edition
// @Description  Create a new edition
// @Tags         Editions
// @Accept       json
// @Produce      json
// @Param        edition  body      model.Edition  true  "Edition to create"
// @Security 	 BearerAuth
// @Success      200  {object}  model.Edition
// @Router       /editions [post]
func (h *Handlers) CreateEdition(r *http.Request) (any, error) {
	var ed model.Edition
	if err := DecodeBody(r, &ed); err != nil {
		return nil, err
	}
	user := r.Context().Value(httpwrapper.GitHubUserKey)
	userLogin := ""
	if u, ok := user.(*httpwrapper.GitHubUser); ok && u != nil {
		userLogin = u.Login
	}
	return h.deps.EditionSvc.CreateEdition(&ed, userLogin)
}

// UpdateEdition godoc
// @Summary      Update Edition
// @Description  Update an existing edition identified by key. The edition data is provided in the JSON body.
// @Tags         Editions
// @Accept       json
// @Produce      json
// @Param        editionId     path      string  true  "Edition ID"
// @Param        edition  body      model.Edition  true  "Edition data to update"
// @Security 	 BearerAuth
// @Success      200  {object}  model.Edition
// @Router       /editions/{editionId} [put]
func (h *Handlers) UpdateEdition(r *http.Request) (any, error) {
	editionId, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}
	var ed model.Edition
	if err := DecodeBody(r, &ed); err != nil {
		return nil, err
	}
	if ed.Key != editionId {
		return nil, fmt.Errorf("edition id in body (%s) does not match id in path (%s)", ed.Key, editionId)
	}
	user := r.Context().Value(httpwrapper.GitHubUserKey)
	userLogin := ""
	if u, ok := user.(*httpwrapper.GitHubUser); ok && u != nil {
		userLogin = u.Login
	}
	return h.deps.EditionSvc.UpdateEdition(&ed, userLogin)
}

// CreateEditionNote godoc
// @Summary      Update Edition Notes
// @Description  Update the notes for an edition identified by id. The note content is provided in the JSON body.
// @Tags         Editions
// @Accept       json
// @Produce      json
// @Param        editionId     path      string  true  "Edition ID"
// @Param        note    body      model.Note  true  "Note content"
// @Security 	 BearerAuth
// @Success      200  {object}  model.Edition
// @Router       /editions/{editionId}/notes [post]
func (h *Handlers) CreateEditionNote(r *http.Request) (any, error) {
	key, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}
	var body model.Note
	if err := DecodeBody(r, &body); err != nil {
		return nil, err
	}
	return h.deps.EditionSvc.UpdateNotes(key, body.Note)
}

// DeleteEdition godoc
// @Summary      Delete Edition
// @Description  Delete an edition identified by ID.
// @Tags         Editions
// @Produce      json
// @Param        editionId  path  string  true  "Edition ID"
// @Security 	 BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /editions/{editionId} [delete]
func (h *Handlers) DeleteEdition(request *http.Request) (any, error) {
	key, err := extractEditionId(request)
	if err != nil {
		return nil, err
	}
	if err := h.deps.EditionSvc.DeleteEdition(key); err != nil {
		return nil, err
	}
	return map[string]any{"success": true, "key": key}, nil
}
