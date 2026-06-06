package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// ListFeatureRevisions godoc
// @Summary      List Edition Feature Revisions
// @Description  Get a list of revisions for a specific edition feature
// @Tags         Edition Feature Revisions
// @Param        featureId     path      string  true  "Feature ID"
// @Produce      json
// @Success      200  {array}   feature.Revision
// @Router        /features/{featureId}/revisions [get]
func (h *Handlers) ListFeatureRevisions(r *http.Request) (any, error) {
	featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	return h.deps.FeatureRevisionSvc.ListFeatureRevisions(featureId)
}

// CreateFeatureRevision godoc
// @Summary      Create Edition Feature Revision
// @Description  Create a new revision for a specific edition feature
// @Tags         Edition Feature Revisions
// @Param        featureId     path      string  true  "Feature ID"
// @Param        revision      body      feature.Revision  true  "Revision data"
// @Produce      json
// @Success      200  {object}  feature.Revision
// @Security 	 BearerAuth
// @Router        /features/{featureId}/revisions [post]
func (h *Handlers) CreateFeatureRevision(r *http.Request) (any, error) {
	featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	var rev feature.Revision
	if err := DecodeBody(r, &rev); err != nil {
		return nil, err
	}
	created, err := h.deps.FeatureRevisionSvc.CreateFeatureRevision(featureId, &rev)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// GetFeatureRevision godoc
// @Summary      Get Edition Feature Revision
// @Description  Get details of a specific edition feature revision
// @Tags         Edition Feature Revisions
// @Param        featureId     path      string  true  "Feature ID"
// @Param        revisionId    path      string  true  "Revision ID"
// @Produce      json
// @Success      200  {object}  feature.Revision
// @Router        /features/{featureId}/revisions/{revisionId} [get]
func (h *Handlers) GetFeatureRevision(r *http.Request) (any, error) {
	featureId, revisionId, err := extractFeatureRevisionID(r)
	if err != nil {
		return nil, err
	}
	return h.deps.FeatureRevisionSvc.GetFeatureRevision(featureId, revisionId)
}
