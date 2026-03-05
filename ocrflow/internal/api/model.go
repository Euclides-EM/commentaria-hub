package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/samber/lo"
)

// ListModels godoc
// @Summary      List Models
// @Description  Get a list of available models.
// @Tags         Models
// @Param        expand query []string false "Include related entities" Enums(used_in_annotations) collectionFormat(multi)
// @Produce      json
// @Success      200  {array}   model.Model
// @Router       /models [get]
func (h *Handlers) ListModels(r *http.Request) (any, error) {
	expand := r.URL.Query()["expand"]
	models, err := h.deps.ModelSvc.List()
	if err != nil {
		return nil, err
	}
	if lo.Contains(expand, "used_in_annotations") {
		usedInAnnotations, err := h.deps.AnnotationSvc.ListAnnotationsByUsedModels()
		if err != nil {
			return nil, err
		}
		for _, m := range models {
			if annIDs, ok := usedInAnnotations[m.ID]; ok {
				m.UsedInAnnotations = annIDs
			} else {
				m.UsedInAnnotations = []*annotation.Reference{}
			}
		}
	}
	return models, nil
}

// UploadModel godoc
// @Summary      Upload a Model
// @Description  Upload a new model to the system.
// @Tags         Models
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Model file to upload"
// @Param        name  formData  string  false  "Name of the model"
// @Param        description  formData  string  false  "Description of the model"
// @Param        base_annotations  formData  string  false  "Comma-separated list of base annotation IDs in the format <dataset_id>:<annotation_id>"
// @Param        base_model_id  formData  string  false  "ID of the base model this model is derived from"
// @Security 	 BearerAuth
// @Success      200   {object}  model.Model
// @Router       /models [post]
func (h *Handlers) UploadModel(r *http.Request) (any, error) {
	name := r.FormValue("name")
	description := r.FormValue("description")
	baseAnnotationsRaw := r.FormValue("base_annotations")
	baseModelID := r.FormValue("base_model_id")
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var baseAnnotations []*annotation.Reference
	for _, rawAnnID := range strings.Split(baseAnnotationsRaw, ",") {
		trimmed := strings.TrimSpace(rawAnnID)
		if trimmed == "" {
			continue
		}
		ids := strings.Split(trimmed, ":")
		if len(ids) != 2 {
			return nil, fmt.Errorf("invalid annotation id, expected format <dataset_id>:<annotation_id>, got %s", trimmed)
		}
		baseAnnotations = append(baseAnnotations, &annotation.Reference{
			DatasetID: ids[0],
			ID:        ids[1],
		})
	}

	return h.deps.ModelSvc.Upload(file, header.Filename, name, description, baseAnnotations, baseModelID)
}

// DeleteModel godoc
// @Summary      Delete a Model
// @Description  Delete a model by its ID.
// @Tags         Models
// @Param        id   path      string  true  "Model ID"
// @Param        deep  query     bool  false  "If true, also delete the model file from filesystem"
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /models/{id} [delete]
func (h *Handlers) DeleteModel(r *http.Request) (any, error) {
	fsClean, err := strconv.ParseBool(r.FormValue("deep"))
	if err != nil {
		fsClean = false
	}
	id := r.PathValue("id")
	if err := h.deps.ModelSvc.Delete(id, fsClean); err != nil {
		return nil, err
	}
	return map[string]string{"status": "deleted"}, nil
}

// UpdateModel godoc
// @Summary      Update a Model
// @Description  Update an existing model.
// @Tags         Models
// @Param        id   path      string  true  "Model ID"
// @Param        model  body      model.Model  true  "Updated model"
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200  {object}   model.Model
// @Router       /models/{id} [put]
func (h *Handlers) UpdateModel(r *http.Request) (any, error) {
	id := r.PathValue("id")
	var toUpdate model.Model
	if err := DecodeBody(r, &toUpdate); err != nil {
		return nil, err
	}
	return h.deps.ModelSvc.Update(id, &toUpdate)
}
