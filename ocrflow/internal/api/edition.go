package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	"github.com/samber/lo"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// ListEditions godoc
// @Summary      List Editions
// @Description  Get a paginated list of editions. Filter by corpus; use offset/limit for paging.
// @Tags         Editions
// @Param        orderBy query     string  false  "Order by field"            Enums(suggested)
// @Param        corpus  query     string  false  "Filter by corpus (comma-separated)"
// @Param        offset  query     int     false  "Paginated offset"         default(0)
// @Param        limit   query     int     false  "Page size (max 100)"      default(20)
// @Produce      json
// @Success      200  {object}  model.EditionListResult
// @Router       /editions [get]
func (h *Handlers) ListEditions(r *http.Request) (any, error) {
	corpuses := strings.Split(r.URL.Query().Get("corpus"), ",")
	corpuses = lo.Filter(lo.Map(corpuses, func(c string, _ int) string {
		return strings.TrimSpace(c)
	}), func(c string, _ int) bool {
		return c != ""
	})
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	return h.deps.EditionSvc.ListEditions(corpuses, model.ToEditionOrderByOptions(r.URL.Query().Get("orderBy")), offset, limit)
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
