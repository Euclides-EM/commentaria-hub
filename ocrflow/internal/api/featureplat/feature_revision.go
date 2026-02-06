package featureplat

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
)

// ListFeatureRevisions godoc
// @Summary      List Feature Revisions
// @Description  Get a list of revisions for a specific feature
// @Tags         Feature Revisions
// @Param        collectionId  path      string  true  "Collection ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Produce      json
// @Success      200  {array}   featureplat.FeatureRevision
// @Router       /collections/{collectionId}/features/{featureId}/revisions [get]
func (h *Handlers) ListFeatureRevisions(r *http.Request) (any, error) {
	collectionId, featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	return h.deps.FeatureRevisionSvc.ListFeatureRevisions(collectionId, featureId)
}

// CreateFeatureRevision godoc
// @Summary      Create Feature Revision
// @Description  Create a new revision for a specific feature
// @Tags         Feature Revisions
// @Param        collectionId  path      string  true  "Collection ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Param        revision      body      featureplat.FeatureRevision  true  "Revision data"
// @Produce      json
// @Success      200  {object}  featureplat.FeatureRevision
// @Router       /collections/{collectionId}/features/{featureId}/revisions [post]
func (h *Handlers) CreateFeatureRevision(r *http.Request) (any, error) {
	collectionId, featureId, err := extractFeatureID(r)
	if err != nil {
		return nil, err
	}
	var rev featureplat.FeatureRevision
	if err := common.DecodeBody(r, &rev); err != nil {
		return nil, err
	}
	created, err := h.deps.FeatureRevisionSvc.CreateFeatureRevision(collectionId, featureId, &rev)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// GetFeatureRevision godoc
// @Summary      Get Feature Revision
// @Description  Get details of a specific feature revision
// @Tags         Feature Revisions
// @Param        collectionId  path      string  true  "Collection ID"
// @Param        featureId     path      string  true  "Feature ID"
// @Param        revisionId    path      string  true  "Revision ID"
// @Produce      json
// @Success      200  {object}  featureplat.FeatureRevision
// @Router       /collections/{collectionId}/features/{featureId}/revisions/{revisionId} [get]
func (h *Handlers) GetFeatureRevision(r *http.Request) (any, error) {
	collectionId, featureId, revisionId, err := extractFeatureRevisionID(r)
	if err != nil {
		return nil, err
	}
	return h.deps.FeatureRevisionSvc.GetFeatureRevision(collectionId, featureId, revisionId)
}
