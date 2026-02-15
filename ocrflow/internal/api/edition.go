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
// @Summary      Get Edition by key
// @Description  Get a single edition by its key.
// @Tags         Editions
// @Param        key  path      string  true  "Edition key"
// @Produce      json
// @Success      200  {object}  model.Edition
// @Failure      404  "Edition not found"
// @Router       /editions/{key} [get]
func (h *Handlers) GetEdition(r *http.Request) (any, error) {
	key, err := extractKey(r)
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
// @Param        key     path      string  true  "Edition key"
// @Param        edition  body      model.Edition  true  "Edition data to update"
// @Security 	 BearerAuth
// @Success      200  {object}  model.Edition
// @Router       /editions/{key} [put]
func (h *Handlers) UpdateEdition(r *http.Request) (any, error) {
	key, err := extractKey(r)
	if err != nil {
		return nil, err
	}
	var ed model.Edition
	if err := DecodeBody(r, &ed); err != nil {
		return nil, err
	}
	if ed.Key != key {
		return nil, fmt.Errorf("edition key in body (%s) does not match key in path (%s)", ed.Key, key)
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
// @Description  Update the notes for an edition identified by key. The note content is provided in the JSON body.
// @Tags         Editions
// @Accept       json
// @Produce      json
// @Param        key     path      string  true  "Edition key"
// @Param        note    body      model.Note  true  "Note content"
// @Security 	 BearerAuth
// @Success      200  {object}  model.Edition
// @Router       /editions/{key}/notes [post]
func (h *Handlers) CreateEditionNote(r *http.Request) (any, error) {
	key, err := extractKey(r)
	if err != nil {
		return nil, err
	}
	var body model.Note
	if err := DecodeBody(r, &body); err != nil {
		return nil, err
	}
	return h.deps.EditionSvc.UpdateNotes(key, body.Note)
}

func (h *Handlers) DeleteEdition(request *http.Request) (any, error) {
	key, err := extractKey(request)
	if err != nil {
		return nil, err
	}
	if err := h.deps.EditionSvc.DeleteEdition(key); err != nil {
		return nil, err
	}
	return map[string]any{"success": true, "key": key}, nil
}

// ImageUpload godoc
// @Summary      Upload Edition Image
// @Description  Upload an image for a specific edition identified by key. The image file is provided as multipart form data.
// @Tags         Editions
// @Accept       multipart/form-data
// @Produce      json
// @Param        key     formData  string  true  "Edition key"
// @Param        type    formData  string  true  "Type of image (e.g., 'cover', 'facsimile')"
// @Param        file    formData  file    true  "Image file to upload"
// @Security 	 BearerAuth
// @Success      200  {object}  model.ImageUpload
// @Router       /editions/upload-image [post]
func (h *Handlers) ImageUpload(r *http.Request) (any, error) {
	key := r.FormValue("key")
	if key == "" {
		return nil, fmt.Errorf("key is required for image upload")
	}
	typ := r.FormValue("type")
	if typ == "" {
		return nil, fmt.Errorf("type is required for image upload")
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return h.deps.EditionSvc.UploadImage(key, typ, file, header)
}
