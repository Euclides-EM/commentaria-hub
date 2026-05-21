package api

import (
	"net/http"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// ListDatasetFeatures godoc
// @Summary      List Features
// @Description  Get a list of available features for the dataset
// @Tags         Features
// @Param        dataSetId  path      string  true  "Dataset ID"
// @Param expand query []string false "Include related entities" Enums(latest_revision,revisions) collectionFormat(multi)
// @Produce      json
// @Success      200  {array}   feature.Feature
// @Router       /datasets/{dataSetId}/features [get]
func (h *Handlers) ListDatasetFeatures(r *http.Request) (any, error) {
	dataSetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	return h.deps.FeatureSvc.ListFeatures(feature.NewDatasetDefScope(dataSetId), feature.ToExpandOptions(r.URL.Query()["expand"]))
}

// ListFeatures godoc
// @Summary      List Edition Features
// @Description  Get a list of available features for the global editions scope
// @Tags         Edition Features
// @Param expand query []string false "Include related entities" Enums(latest_revision,revisions) collectionFormat(multi)
// @Param scope  query string true "Filter by feature execution scope" Enums(dataset, editions)
// @Param dataset query string false "Filter by dataset ID, relevant only for the dataset scope"
// @Produce      json
// @Success      200  {array}   feature.Feature
// @Router       /features [get]
func (h *Handlers) ListFeatures(r *http.Request) (any, error) {
	scope, err := extractScope(r)
	if err != nil {
		return nil, err
	}
	return h.deps.FeatureSvc.ListFeatures(scope, feature.ToExpandOptions(r.URL.Query()["expand"]))
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
	f.Scope = feature.NewDatasetDefScope(dataSetId)
	created, err := h.deps.FeatureSvc.Create(&f)
	if err != nil {
		return nil, err
	}
	return created, nil
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

// DeleteDatasetFeature godoc
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
func (h *Handlers) DeleteDatasetFeature(r *http.Request) (any, error) {
	dataSetId, featureId, err := extractDatasetFeatureID(r)
	if err != nil {
		return nil, err
	}
	force, err := strconv.ParseBool(r.URL.Query().Get("force"))
	if err != nil {
		force = false // default to false if parsing fails or not provided
	}
	if err = h.deps.FeatureSvc.DeleteFeatureInScope(feature.NewDatasetDefScope(dataSetId), featureId, force); err != nil {
		return nil, err
	}
	return map[string]string{"status": "deleted"}, nil
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

// GetDatasetFeature godoc
// @Summary      Get Feature
// @Description  Get details of a specific feature from the dataset
// @Tags         Features
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Param        expand query []string false "Include related entities" Enums(latest_revision,revisions) collectionFormat(multi)
// @Produce      json
// @Success      200  {object}  feature.Feature
// @Router        /datasets/{dataSetId}/features/{featureId} [get]
func (h *Handlers) GetDatasetFeature(r *http.Request) (any, error) {
	dataSetId, featureId, err := extractDatasetFeatureID(r)
	if featureId == "" {
		return nil, err
	}
	feat, err := h.deps.FeatureSvc.GetFeatureInScope(feature.NewDatasetDefScope(dataSetId), featureId, feature.ToExpandOptions(r.URL.Query()["expand"]))
	if err != nil {
		return nil, err
	}
	return feat, nil
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

// UpdateDatasetFeature godoc
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
func (h *Handlers) UpdateDatasetFeature(r *http.Request) (any, error) {
	dataSetId, featureId, err := extractDatasetFeatureID(r)
	if err != nil {
		return nil, err
	}
	var f feature.Feature
	if err := DecodeBody(r, &f); err != nil {
		return nil, err
	}
	f.Scope = feature.NewDatasetDefScope(dataSetId)
	f.ID = featureId
	updated, err := h.deps.FeatureSvc.UpdateFeature(&f)
	if err != nil {
		return nil, err
	}
	return updated, nil
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
