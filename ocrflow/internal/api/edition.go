package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/search"
)

const (
	defaultListLimit = 20
	maxListLimit     = 5000
	diagramsPathBase = "diagrams"
	diagramsSubDir   = "diagrams"
	diagramsCropsDir = "crops"
)

type editionDiagramsResponse struct {
	Key             string                 `json:"key,omitempty"`
	ImageURLsByName map[string]string      `json:"imageUrlsByName,omitempty"`
	HasNoDiagrams   bool                   `json:"hasNoDiagrams"`
	Volumes         []editionDiagramVolume `json:"volumes,omitempty"`
}

type editionDiagramVolume struct {
	Volume          int               `json:"volume,omitempty"`
	Key             string            `json:"key,omitempty"`
	ImageURLsByName map[string]string `json:"imageUrlsByName"`
	HasNoDiagrams   bool              `json:"hasNoDiagrams"`
}

type editionDiagramsFileData struct {
	Key           string                     `json:"key,omitempty"`
	Images        []string                   `json:"images,omitempty"`
	HasNoDiagrams bool                       `json:"hasNoDiagrams"`
	Volumes       []editionDiagramFileVolume `json:"volumes,omitempty"`
}

type editionDiagramFileVolume struct {
	Volume        int      `json:"volume,omitempty"`
	Key           string   `json:"key,omitempty"`
	Images        []string `json:"images"`
	HasNoDiagrams bool     `json:"hasNoDiagrams"`
}

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

// DeleteEdition godoc
// @Summary      Delete Edition
// @Description  Delete an edition identified by key.
// @Tags         Editions
// @Produce      json
// @Param        key  path  string  true  "Edition key"
// @Security 	 BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  "Missing key"
// @Router       /editions/{key} [delete]
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

// ListEditionDiagrams godoc
// @Summary      List Edition Diagram Directories
// @Description  Get all available edition diagram directory keys.
// @Tags         Editions
// @Produce      json
// @Success      200  {array}  string
// @Router       /editions/diagrams [get]
func (h *Handlers) ListEditionDiagrams(_ *http.Request) (any, error) {
	path := filepath.Join(h.deps.Env.ItemsMetadataStoreDir, "diagram-directories.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read diagram directories: %w", err)
	}

	var directories []string
	if err := json.Unmarshal(data, &directories); err != nil {
		return nil, fmt.Errorf("decode diagram directories: %w", err)
	}
	return directories, nil
}

// GetEditionDiagrams godoc
// @Summary      Get Edition Diagrams
// @Description  Get diagram image URLs for a specific edition key.
// @Tags         Editions
// @Produce      json
// @Param        key  path  string  true  "Edition key"
// @Success      200  {object}  editionDiagramsResponse
// @Failure      400  "Missing key"
// @Router       /editions/{key}/diagrams [get]
func (h *Handlers) GetEditionDiagrams(r *http.Request) (any, error) {
	key, err := extractKey(r)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(h.deps.Env.ItemsMetadataStoreDir, diagramsSubDir, key+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return editionDiagramsResponse{
				ImageURLsByName: map[string]string{},
				HasNoDiagrams:   true,
			}, nil
		}
		return nil, fmt.Errorf("read diagrams for %s: %w", key, err)
	}

	var fileData editionDiagramsFileData
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("decode diagrams for %s: %w", key, err)
	}
	diagramsContentBase := resolveDiagramsContentBase(h.deps.Env.FacsimilesGithubRepoUrl)
	response := editionDiagramsResponse{
		Key:             fileData.Key,
		HasNoDiagrams:   fileData.HasNoDiagrams,
		ImageURLsByName: map[string]string{},
	}

	if len(fileData.Volumes) > 0 {
		response.Volumes = make([]editionDiagramVolume, 0, len(fileData.Volumes))
		for i := range fileData.Volumes {
			volumeKey := fileData.Volumes[i].Key
			if volumeKey == "" {
				volumeKey = key
			}
			response.Volumes = append(response.Volumes, editionDiagramVolume{
				Volume:        fileData.Volumes[i].Volume,
				Key:           fileData.Volumes[i].Key,
				HasNoDiagrams: fileData.Volumes[i].HasNoDiagrams,
				ImageURLsByName: mapDiagramImageURLsByName(
					diagramsContentBase,
					volumeKey,
					fileData.Volumes[i].Images,
				),
			})
		}
		return response, nil
	}

	singleKey := fileData.Key
	if singleKey == "" {
		singleKey = key
	}
	response.ImageURLsByName = mapDiagramImageURLsByName(
		fmt.Sprintf("%s/raw/main/docs", h.deps.Env.FacsimilesGithubRepoUrl),
		singleKey,
		fileData.Images,
	)
	return response, nil
}

func mapDiagramImageURLsByName(baseURL, key string, images []string) map[string]string {
	out := make(map[string]string, len(images))
	for _, imageName := range images {
		out[imageName] = buildDiagramImageURL(baseURL, key, imageName)
	}
	return out
}

func buildDiagramImageURL(baseURL, key, imageName string) string {
	return fmt.Sprintf(
		"%s/diagrams/%s/%s/%s",
		baseURL,
		url.PathEscape(key),
		diagramsCropsDir,
		url.PathEscape(imageName),
	)
}

func resolveDiagramsContentBase(facsimilesRepoURL string) string {
	repoURL := strings.TrimSuffix(strings.TrimSpace(facsimilesRepoURL), "/")
	parsed, err := url.Parse(repoURL)
	if err != nil {
		return repoURL
	}

	if strings.EqualFold(parsed.Host, "raw.githubusercontent.com") {
		return repoURL
	}

	if !strings.EqualFold(parsed.Host, "github.com") {
		return repoURL
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return repoURL
	}

	contentPath := append([]string{parts[0], parts[1], "blob", "main", "docs"})
	return "https://raw.githubusercontent.com/" + strings.Join(contentPath, "/")
}
