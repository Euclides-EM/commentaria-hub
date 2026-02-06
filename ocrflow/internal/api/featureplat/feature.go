package featureplat

import (
	"net/http"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
)

// ListFeatures godoc
// @Summary      List Features
// @Description  Get a list of available features for the collection
// @Tags         Features
// @Param        collectionId  path      string  true  "Collection ID"
// @Param expand query []string false "Include related entities" Enums(latest_revision,revisions) collectionFormat(multi)
// @Produce      json
// @Success      200  {array}   featureplat.Feature
// @Router       /collections/{collectionId}/features [get]
func (h *Handlers) ListFeatures(r *http.Request) (any, error) {
	return h.deps.FeatureSvc.ListFeatures(featureplat.ToFeatureExpandOptions(r.URL.Query().Get("expand")))
}

// CreateFeatures godoc
// @Summary      Create Feature
// @Description  Create a new feature for the collection
// @Tags         Features
// @Accept       json
// @Produce      json
// @Param        collectionId  path      string  true  "Collection ID"
// @Param        feature  body      featureplat.Feature  true  "Feature to create"
// @Success      200  {object}  featureplat.Feature
// @Router       /collections/{collectionId}/features [post]
func (h *Handlers) CreateFeatures(r *http.Request) (any, error) {
	var f featureplat.Feature
	if err := common.DecodeBody(r, &f); err != nil {
		return nil, err
	}
	created, err := h.deps.FeatureSvc.CreateFeature(&f)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// DeleteFeature godoc
// @Summary      Delete Feature
// @Description  Delete a feature from the collection
// @Tags         Features
// @Param        collectionId  path      string  true  "Collection ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Produce      json
// @Success      204  "No Content"
// @Router       /collections/{collectionId}/features/{featureId} [delete]
func (h *Handlers) DeleteFeature(r *http.Request) (any, error) {
	collectionId, featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	force, err := strconv.ParseBool(r.URL.Query().Get("force"))
	if err != nil {
		force = false // default to false if parsing fails or not provided
	}
	if err = h.deps.FeatureSvc.Delete(collectionId, featureId, force); err != nil {
		return nil, err
	}
	return map[string]string{"status": "deleted"}, nil
}

// GetFeature godoc
// @Summary      Get Feature
// @Description  Get details of a specific feature from the collection
// @Tags         Features
// @Param        collectionId  path      string  true  "Collection ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Param expand query []string false "Include related entities" Enums(latest_revision,revisions) collectionFormat(multi)
// @Produce      json
// @Success      200  {object}  featureplat.Feature
// @Router       /collections/{collectionId}/features/{featureId} [get]
func (h *Handlers) GetFeature(r *http.Request) (any, error) {
	collectionId, featureId, err := extractFeatureID(r)
	if featureId == "" {
		return nil, err
	}
	feat, err := h.deps.FeatureSvc.GetFeature(collectionId, featureId, featureplat.ToFeatureExpandOptions(r.URL.Query().Get("expand")))
	if err != nil {
		return nil, err
	}
	return feat, nil
}

// UpdateFeature godoc
// @Summary      Update Feature
// @Description  Update an existing feature in the collection
// @Tags         Features
// @Param        collectionId  path      string  true  "Collection ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Param        feature       body      featureplat.Feature  true  "Updated feature data"
// @Produce      json
// @Success      200  {object}  featureplat.Feature
// @Router       /collections/{collectionId}/features/{featureId} [put]
func (h *Handlers) UpdateFeature(r *http.Request) (any, error) {
	collectionId, featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	var f featureplat.Feature
	if err := common.DecodeBody(r, &f); err != nil {
		return nil, err
	}
	updated, err := h.deps.FeatureSvc.UpdateFeature(collectionId, featureId, &f)
	if err != nil {
		return nil, err
	}
	return updated, nil
}
