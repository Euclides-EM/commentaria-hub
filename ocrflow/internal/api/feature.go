package api

import (
	"net/http"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// ListFeatures godoc
// @Summary      List Features
// @Description  Get a list of available features for the dataset
// @Tags         Features
// @Param        dataSetId  path      string  true  "Dataset ID"
// @Param expand query []string false "Include related entities" Enums(latest_revision,revisions) collectionFormat(multi)
// @Produce      json
// @Success      200  {array}   feature.Feature
// @Router       /datasets/{dataSetId}/features [get]
func (h *Handlers) ListFeatures(r *http.Request) (any, error) {
	dataSetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	return h.deps.FeatureSvc.ListFeatures(dataSetId, feature.ToFeatureExpandOptions(r.URL.Query()["expand"]))
}

// CreateFeatures godoc
// @Summary      Create Feature
// @Description  Create a new feature for the dataset
// @Tags         Features
// @Accept       json
// @Produce      json
// @Param        dataSetId path      string  true  "Dataset ID"
// @Param        feature  body      feature.Feature  true  "Feature to create"
// @Success      200  {object}  feature.Feature
// @Security 	 BearerAuth
// @Router        /datasets/{dataSetId}/features [post]
func (h *Handlers) CreateFeatures(r *http.Request) (any, error) {
	dataSetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	var f feature.Feature
	if err := DecodeBody(r, &f); err != nil {
		return nil, err
	}
	created, err := h.deps.FeatureSvc.CreateFeature(dataSetId, &f)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// DeleteFeature godoc
// @Summary      Delete Feature
// @Description  Delete a feature from the dataset.
// @Tags         Features
// @Param        dataSetId    path      string  true  "Dataset ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Param        force         query     bool  false "Force deletion"
// @Produce      json
// @Success      204  "No Content"
// @Security 	 BearerAuth
// @Router        /datasets/{dataSetId}/features/{featureId} [delete]
func (h *Handlers) DeleteFeature(r *http.Request) (any, error) {
	dataSetId, featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	force, err := strconv.ParseBool(r.URL.Query().Get("force"))
	if err != nil {
		force = false // default to false if parsing fails or not provided
	}
	if err = h.deps.FeatureSvc.Delete(dataSetId, featureId, force); err != nil {
		return nil, err
	}
	return map[string]string{"status": "deleted"}, nil
}

// GetFeature godoc
// @Summary      Get Feature
// @Description  Get details of a specific feature from the dataset
// @Tags         Features
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Param        expand query []string false "Include related entities" Enums(latest_revision,revisions) collectionFormat(multi)
// @Produce      json
// @Success      200  {object}  feature.Feature
// @Router        /datasets/{dataSetId}/features/{featureId} [get]
func (h *Handlers) GetFeature(r *http.Request) (any, error) {
	dataSetId, featureId, err := extractFeatureID(r)
	if featureId == "" {
		return nil, err
	}
	feat, err := h.deps.FeatureSvc.GetFeature(dataSetId, featureId, feature.ToFeatureExpandOptions(r.URL.Query()["expand"]))
	if err != nil {
		return nil, err
	}
	return feat, nil
}

// UpdateFeature godoc
// @Summary      Update Feature
// @Description  Update an existing feature in the dataset.
// @Tags         Features
// @Param        dataSetId  path      string  true  "Dataset ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Param        feature       body      feature.Feature  true  "Updated feature data"
// @Produce      json
// @Success      200  {object}  feature.Feature
// @Security 	 BearerAuth
// @Router        /datasets/{dataSetId}/features/{featureId} [put]
func (h *Handlers) UpdateFeature(r *http.Request) (any, error) {
	dataSetId, featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	var f feature.Feature
	if err := DecodeBody(r, &f); err != nil {
		return nil, err
	}
	updated, err := h.deps.FeatureSvc.UpdateFeature(dataSetId, featureId, &f)
	if err != nil {
		return nil, err
	}
	return updated, nil
}
