package api

import (
	"net/http"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// ListFeatures godoc
// @Summary      List Edition Features
// @Description  Get a list of available features for the global editions scope
// @Tags         Edition Features
// @Param expand query []string false "Include related entities" Enums(latest_revision,revisions) collectionFormat(multi)
// @Param scope  query string true "Filter by feature execution scope" Enums(dataset, editions)
// @Param dataset query string false "Filter by dataset ID, relevant only for the dataset scope; if omitted, returns features from all datasets"
// @Produce      json
// @Success      200  {array}   feature.Feature
// @Router       /features [get]
func (h *Handlers) ListFeatures(r *http.Request) (any, error) {
	scope, err := extractDefScope(r)
	if err != nil {
		return nil, err
	}
	return h.deps.FeatureSvc.ListFeatures(scope, feature.ToExpandOptions(r.URL.Query()["expand"]))
}

// CreateFeature godoc
// @Summary      Create Edition Feature
// @Description  Create a new feature for the global editions scope
// @Tags         Edition Features
// @Accept       json
// @Produce      json
// @Param        feature  body      feature.Feature  true  "Feature to create"
// @Success      200  {object}  feature.Feature
// @Security 	 BearerAuth
// @Router       /features [post]
func (h *Handlers) CreateFeature(r *http.Request) (any, error) {
	var f feature.Feature
	if err := DecodeBody(r, &f); err != nil {
		return nil, err
	}
	created, err := h.deps.FeatureSvc.Create(&f)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// DeleteFeature godoc
// @Summary      Delete Edition Feature
// @Description  Delete a feature from the global editions scope.
// @Tags         Edition Features
// @Param        featureId     path      string  true  "Feature ID"
// @Param        force         query     bool  false "Force deletion"
// @Produce      json
// @Success      204  "No Content"
// @Security 	 BearerAuth
// @Router       /features/{featureId} [delete]
func (h *Handlers) DeleteFeature(r *http.Request) (any, error) {
	featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	force, err := strconv.ParseBool(r.URL.Query().Get("force"))
	if err != nil {
		force = false
	}
	if err = h.deps.FeatureSvc.DeleteFeature(featureId, force); err != nil {
		return nil, err
	}
	return map[string]string{"status": "deleted"}, nil
}

// DeleteFeatures godoc
// @Summary      Delete Edition Features
// @Description  Delete multiple features in a scope.
// @Tags         Edition Features
// @Param        scope         query     string  false "Feature scope" Enums(dataset, editions)
// @Param        dataset       query     string  false "Dataset ID, relevant only for the dataset scope"
// @Param        ids           query     []string  false "feature IDs" collectionFormat(multi)
// @Produce      json
// @Success      200  {object}  map[string]any
// @Security 	 BearerAuth
// @Router       /features [delete]
func (h *Handlers) DeleteFeatures(r *http.Request) (any, error) {
	ids := r.URL.Query()["ids"]
	scope, err := extractDefScope(r)
	if err != nil {
		return nil, err
	}
	deleted, err := h.deps.FeatureSvc.DeleteFeaturesInScope(scope, ids, false)
	if err != nil {
		return nil, err
	}
	return map[string]any{"status": "deleted", "deleted": deleted}, nil
}

// GetFeature godoc
// @Summary      Get Edition Feature
// @Description  Get details of a specific feature from the global editions scope
// @Tags         Edition Features
// @Param        featureId     path      string  true  "Feature ID"
// @Param        expand query []string false "Include related entities" Enums(latest_revision,revisions) collectionFormat(multi)
// @Produce      json
// @Success      200  {object}  feature.Feature
// @Router        /features/{featureId} [get]
func (h *Handlers) GetFeature(r *http.Request) (any, error) {
	featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	feat, err := h.deps.FeatureSvc.GetFeature(featureId, feature.ToExpandOptions(r.URL.Query()["expand"]))
	if err != nil {
		return nil, err
	}
	return feat, nil
}

// UpdateEditionsFeature godoc
// @Summary      Update Edition Feature
// @Description  Update an existing feature in the global editions scope.
// @Tags         Edition Features
// @Param        featureId     path      string  true  "Feature ID"
// @Param        feature       body      feature.Feature  true  "Updated feature data"
// @Produce      json
// @Success      200  {object}  feature.Feature
// @Security 	 BearerAuth
// @Router        /features/{featureId} [put]
func (h *Handlers) UpdateEditionsFeature(r *http.Request) (any, error) {
	featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	var f feature.Feature
	if err := DecodeBody(r, &f); err != nil {
		return nil, err
	}
	f.ID = featureId
	updated, err := h.deps.FeatureSvc.UpdateFeature(&f)
	if err != nil {
		return nil, err
	}
	return updated, nil
}
