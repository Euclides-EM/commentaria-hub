package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotationrule"
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
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply [put]
func (h *Handlers) ApplyRules(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	var a annotationrule.ApplyRules
	if err = DecodeBody(r, &a); err != nil {
		return nil, err
	}

	return h.deps.AnnotationRuleExecution.ApplyRules(datasetID, annotationID, &a)
}

// ApplyRuleModelDetect godoc
// @Summary      Apply ModelDetect Rule to Annotation
// @Description  Apply a model detect rule to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationModelDetectRule  body 	annotationrule.ModelDetect  true  "Annotation model detect rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/model_detect [put]
func (h *Handlers) ApplyRuleModelDetect(r *http.Request) (any, error) {
	var rule annotationrule.ModelDetect
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleSlicePages godoc
// @Summary      Apply Slice Pages Rule to Annotation
// @Description  Apply a slice pages rule to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.SlicePages  true  "Annotation slice pages rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/slice_pages [put]
func (h *Handlers) ApplyRuleSlicePages(r *http.Request) (any, error) {
	var rule annotationrule.SlicePages
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleStretch godoc
// @Summary      Apply Stretch Rule to Annotation
// @Description  Apply a stretch rule to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.Stretch  true  "Annotation stretch rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/stretch [put]
func (h *Handlers) ApplyRuleStretch(r *http.Request) (any, error) {
	var rule annotationrule.Stretch
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleAddMargin godoc
// @Summary      Add Margin Rule to Annotation
// @Description  add margin to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.AddMargin  true  "Annotation add margin rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/add_margin [put]
func (h *Handlers) ApplyRuleAddMargin(r *http.Request) (any, error) {
	var rule annotationrule.AddMargin
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleDetectLines godoc
// @Summary      Detect Lines in Annotation
// @Description  Detect lines in an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.LinesDetect  true  "Annotation detect lines rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/detect_lines [put]
func (h *Handlers) ApplyRuleDetectLines(r *http.Request) (any, error) {
	var rule annotationrule.LinesDetect
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleRemoveCategories godoc
// @Summary      Remove Categories in Annotation
// @Description  Remove categories in an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.RemoveCategories  true  "Remove categories rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/remove_categories [put]
func (h *Handlers) ApplyRuleRemoveCategories(r *http.Request) (any, error) {
	var rule annotationrule.RemoveCategories
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleRenameCategories godoc
// @Summary      Rename Categories in Annotation
// @Description  Rename categories in an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.RenameCategories  true  "Rename categories rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/rename_categories [put]
func (h *Handlers) ApplyRuleRenameCategories(r *http.Request) (any, error) {
	var rule annotationrule.RenameCategories
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleRemoveOverlap godoc
// @Summary      Remove Overlap in Annotation
// @Description  Remove overlapping annotations in an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.RemoveOverlap  true  "Remove overlap rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/remove_overlap [put]
func (h *Handlers) ApplyRuleRemoveOverlap(r *http.Request) (any, error) {
	var rule annotationrule.RemoveOverlap
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleResolveOverlapWithPriority godoc
// @Summary      Resolve Overlap with Priority in Annotation
// @Description  Remove zones of SuppressedCategory that overlap DominantCategory by at least MinOverlap percent.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.ResolveOverlapWithPriority  true  "Resolve overlap with priority rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/resolve_overlap_with_priority [put]
func (h *Handlers) ApplyRuleResolveOverlapWithPriority(r *http.Request) (any, error) {
	var rule annotationrule.ResolveOverlapWithPriority
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleRecategorizeByAlignment godoc
// @Summary      Recategorize by Alignment in Annotation
// @Description  Recategorize zones that are horizontally or vertically aligned with zones of another category (within pixel tolerance).
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.RecategorizeByAlignment  true  "Recategorize by alignment rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/recategorize_by_alignment [put]
func (h *Handlers) ApplyRuleRecategorizeByAlignment(r *http.Request) (any, error) {
	var rule annotationrule.RecategorizeByAlignment
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleLimitCategoryZones godoc
// @Summary      Limit Category Zones in Annotation
// @Description  Keep at most max_count zones of a category per page, keeping those closest to the specified page side (top/bottom/left/right).
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.LimitCategoryZones  true  "Limit category zones rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/limit_category_zones [put]
func (h *Handlers) ApplyRuleLimitCategoryZones(r *http.Request) (any, error) {
	var rule annotationrule.LimitCategoryZones
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleReassignTextLinesByTolerance godoc
// @Summary      Reassign Text Lines by Tolerance in Annotation
// @Description  Reassign text lines by tolerance in an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationRule  body 	annotationrule.ReassignTextLinesByTolerance  true  "Reassign text lines by tolerance rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/reassign_text_lines_by_tolerance [put]
func (h *Handlers) ApplyRuleReassignTextLinesByTolerance(r *http.Request) (any, error) {
	var rule annotationrule.ReassignTextLinesByTolerance
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

// ApplyRuleTextBlockCorrections godoc
// @Summary      Apply Text Block Corrections to Annotation
// @Description  Apply text block corrections to an annotation.
// @Tags         Annotations Apply Rules
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        action      query     string  false "Action to take when applying the rule" Enums(overwrite,create_new) default(overwrite)
// @Param        execution_mode query  string  false "Execution mode for applying the rule" Enums(sync,async) default(sync)
// @Param        annotationTextBlockCorrections  body 	annotationrule.TextBlockCorrections  true  "Text block corrections rule"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/text_block_corrections [put]
func (h *Handlers) ApplyRuleTextBlockCorrections(r *http.Request) (any, error) {
	var rule annotationrule.TextBlockCorrections
	if err := DecodeBody(r, &rule); err != nil {
		return nil, err
	}
	rule.Type = rule.GetType()
	rule.ApplicableStages = annotationrule.GetApplicableStages(rule.GetType())

	return h.applyRuleGeneric(r, &rule)
}

func (h *Handlers) applyRuleGeneric(r *http.Request, rule annotationrule.AnnotationRule) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	rules := &annotationrule.ApplyRules{
		Action:        annotationrule.ToApplyRulesAction(r.FormValue("action"), annotationrule.ApplyRulesActionOverwrite),
		ExecutionMode: annotationrule.ToExecutionMode(r.FormValue("execution_mode"), annotationrule.ExecutionModeSync),
		Rules: []annotationrule.AnnotationRule{
			rule,
		},
	}

	return h.deps.AnnotationRuleExecution.ApplyRules(datasetID, annotationID, rules)
}
