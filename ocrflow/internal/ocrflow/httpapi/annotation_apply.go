package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
)

// ApplyRules godoc
// @Summary      Apply Rules to Annotation
// @Description  Apply specific rules to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationApplyRules  body 	annotationrule.ApplyRules  true  "Annotation apply rules"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply [put]
func (h *Handlers) ApplyRules(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")

	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var a annotationrule.ApplyRules
	if err := decoder.Decode(&a); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, &a)
}

// ApplyRuleSegment godoc
// @Summary      Apply Segment Rule to Annotation
// @Description  Apply a segment rule to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.Segment  true  "Annotation segment rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/segment [put]
func (h *Handlers) ApplyRuleSegment(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.Segment
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleSlicePages godoc
// @Summary      Apply Slice Pages Rule to Annotation
// @Description  Apply a slice pages rule to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.SlicePages  true  "Annotation slice pages rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/slice_pages [put]
func (h *Handlers) ApplyRuleSlicePages(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.SlicePages
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleStretch godoc
// @Summary      Apply Stretch Rule to Annotation
// @Description  Apply a stretch rule to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.Stretch  true  "Annotation stretch rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/stretch [put]
func (h *Handlers) ApplyRuleStretch(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.Stretch
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleAddMargin godoc
// @Summary      Add Margin Rule to Annotation
// @Description  add margin to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.AddMargin  true  "Annotation add margin rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/add_margin [put]
func (h *Handlers) ApplyRuleAddMargin(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.AddMargin
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleDetectLines godoc
// @Summary      Detect Lines in Annotation
// @Description  Detect lines in an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.LinesDetect  true  "Annotation detect lines rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/detect_lines [put]
func (h *Handlers) ApplyRuleDetectLines(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.LinesDetect
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleRemoveCategories godoc
// @Summary      Remove Categories in Annotation
// @Description  Remove categories in an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.RemoveCategories  true  "Remove categories rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/remove_categories [put]
func (h *Handlers) ApplyRuleRemoveCategories(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.RemoveCategories
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleRemoveOverlap godoc
// @Summary      Remove Overlap in Annotation
// @Description  Remove overlapping annotations in an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.RemoveOverlap  true  "Remove overlap rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/remove_overlap [put]
func (h *Handlers) ApplyRuleRemoveOverlap(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.RemoveOverlap
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleReassignTextLinesByTolerance godoc
// @Summary      Reassign Text Lines by Tolerance in Annotation
// @Description  Reassign text lines by tolerance in an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.ReassignTextLinesByTolerance  true  "Reassign text lines by tolerance rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/reassign_text_lines_by_tolerance [put]
func (h *Handlers) ApplyRuleReassignTextLinesByTolerance(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.ReassignTextLinesByTolerance
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}
