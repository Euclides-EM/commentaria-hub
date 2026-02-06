package ocrflow

import "net/http"

// ListAnnotationRules godoc
// @Summary      List Annotation Rules
// @Description  Lists all available annotation rules with their default configurations.
// @Tags         Metadata
// @Produce      json
// @Success      200  {array}   annotationrule.MetadataDetails
// @Router       /annotation_rules [get]
func (h *Handlers) ListAnnotationRules(r *http.Request) (any, error) {
	return h.deps.MetadataDetailsSvc.ListAnnotationRules()
}

// ListPipelineStages godoc
// @Summary      List Pipeline Stages
// @Description  Lists all defined pipeline stages for annotations.
// @Tags         Metadata
// @Produce      json
// @Success      200  {array}   annotationrule.PipelineStage
// @Router       /pipeline_stages [get]
func (h *Handlers) ListPipelineStages(r *http.Request) (any, error) {
	return h.deps.MetadataDetailsSvc.ListPipelineStages()
}
