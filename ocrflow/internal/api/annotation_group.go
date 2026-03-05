package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
)

// ListAnnotationGroups godoc
// @Summary      List Annotation Groups
// @Description  Get a list of all annotation groups
// @Tags         Annotation Groups
// @Produce      json
// @Success      200  {array}   annotation.Group
// @Router        /annotation-groups [get]
func (h *Handlers) ListAnnotationGroups(request *http.Request) (any, error) {
	return h.deps.AnnotationGroupSvc.List()
}

// GetAnnotationGroup godoc
// @Summary      Get Annotation Group
// @Description  Get details of a specific annotation group by ID
// @Tags         Annotation Groups
// @Param        groupId     path      string  true  "Annotation Group ID"
// @Produce      json
// @Success      200  {object}  annotation.Group
// @Failure      404  {string}  string "Not Found"
// @Router        /annotation-groups/{groupId} [get]
func (h *Handlers) GetAnnotationGroup(request *http.Request) (any, error) {
	groupId, err := extractGroupId(request)
	if err != nil {
		return nil, err
	}
	return h.deps.AnnotationGroupSvc.Get(groupId)
}

// CreateAnnotationGroup godoc
// @Summary      Create Annotation Group
// @Description  Create a new annotation group with the provided details
// @Tags         Annotation Groups
// @Param        group       body      annotation.GroupCreateRequest  true  "Annotation Group data"
// @Produce      json
// @Success      200  {object}  annotation.Group
// @Security 	 BearerAuth
// @Router        /annotation-groups [post]
func (h *Handlers) CreateAnnotationGroup(request *http.Request) (any, error) {
	annGroup := &annotation.Group{}
	err := DecodeBody(request, annGroup)
	if err != nil {
		return nil, err
	}
	return h.deps.AnnotationGroupSvc.Create(annGroup)
}

// UpdateAnnotationGroup godoc
// @Summary      Update Annotation Group
// @Description  Update details of an existing annotation group by ID
// @Tags         Annotation Groups
// @Param        groupId     path      string  true  "Annotation Group ID"
// @Param        group       body      annotation.GroupUpdateRequest  true  "Updated Annotation Group data"
// @Produce      json
// @Success      200  {object}  annotation.Group
// @Failure      404  {string}  string "Not Found"
// @Security 	 BearerAuth
// @Router        /annotation-groups/{groupId} [put]
func (h *Handlers) UpdateAnnotationGroup(request *http.Request) (any, error) {
	groupId, err := extractGroupId(request)
	if err != nil {
		return nil, err
	}
	annGroup := &annotation.Group{}
	err = DecodeBody(request, annGroup)
	if err != nil {
		return nil, err
	}
	return h.deps.AnnotationGroupSvc.Update(groupId, annGroup)
}

// DeleteAnnotationGroup godoc
// @Summary      Delete Annotation Group
// @Description  Delete an annotation group by ID
// @Tags         Annotation Groups
// @Param        groupId     path      string  true  "Annotation Group ID"
// @Produce      json
// @Success      204  {string}  string "No Content"
// @Failure      404  {string}  string "Not Found"
// @Security 	 BearerAuth
// @Router        /annotation-groups/{groupId} [delete]
func (h *Handlers) DeleteAnnotationGroup(request *http.Request) (any, error) {
	groupId, err := extractGroupId(request)
	if err != nil {
		return nil, err
	}
	return nil, h.deps.AnnotationGroupSvc.Delete(groupId)
}
