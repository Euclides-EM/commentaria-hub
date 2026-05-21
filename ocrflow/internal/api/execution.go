package api

import (
	"fmt"
	"net/http"
	"strings"

	mfeatureplat "github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/samber/lo"
)

// ListExecutions godoc
// @Summary      List Executions
// @Description  Get a list of all executions for a specific edition
// @Tags         Executions
// @Produce      json
// @Success      200  {array}   feature.Execution
// @Param dataset query string false "Filter by dataset ID"
// @Param scope query string false "Filter by feature execution scope" Enums(dataset, editions)
// @Param features query string false "Filter by delimited list of feature IDs"
// @Param statuses query string false "Filter by delimited list of execution statuses" Enums(pending, running, completed, failed)
// @Router       /feature_executions [get]
func (h *Handlers) ListExecutions(r *http.Request) (any, error) {
	scope, err := extractScope(r)
	if err != nil {
		return nil, err
	}
	features := r.URL.Query().Get("features")
	statuses := r.URL.Query().Get("statuses")

	var featureIds []string
	fs, err := h.deps.FeatureSvc.ListFeatures(scope, nil)
	if err != nil {
		return nil, err
	}

	if scope.DatasetID != "" && scope.Type == mfeatureplat.ScopeTypeDataset {
		fs = lo.Filter(fs, func(f *mfeatureplat.Feature, _ int) bool {
			return f.Scope.DatasetID == scope.DatasetID
		})
	}

	if features != "" {
		rawFeatureIds := lo.Map(strings.Split(features, ","), func(s string, _ int) string {
			return strings.TrimSpace(s)
		})
		fs = lo.Filter(fs, func(f *mfeatureplat.Feature, _ int) bool {
			return lo.Contains(rawFeatureIds, f.ID)
		})
	}
	for _, f := range fs {
		featureIds = append(featureIds, f.ID)
	}

	var featureExecutionsStatuses []mfeatureplat.ExecutionStatus
	if statuses != "" {
		for _, s := range strings.Split(statuses, ",") {
			fxs := mfeatureplat.ToExecutionsStatus(s)
			if fxs == "" {
				return nil, fmt.Errorf("invalid execution status: %s", s)
			}
			featureExecutionsStatuses = append(featureExecutionsStatuses, fxs)
		}
	}

	return h.deps.FeatureExecutionSvc.ListFeatureExecutions(scope, featureIds, featureExecutionsStatuses)
}

// GetExecution godoc
// @Summary      Get Execution
// @Description  Get details of a specific execution by ID
// @Tags         Executions
// @Produce      json
// @Success      200  {object}  feature.Execution
// @Param executionId path string true "Execution ID"
// @Router       /feature_executions/{executionId} [get]
func (h *Handlers) GetExecution(r *http.Request) (any, error) {
	executionId, err := extractExecutionID(r)
	if err != nil {
		return nil, err
	}
	return h.deps.FeatureExecutionSvc.GetFeatureExecution(executionId)
}

// CreateExecution godoc
// @Summary      Create Execution
// @Description  Create a new execution for a specific feature
// @Tags         Executions
// @Accept       json
// @Produce      json
// @Param        execution  body      feature.Execution  true  "Execution to create"
// @Security 	 BearerAuth
// @Success      200  {object}  feature.Execution
// @Router       /feature_executions [post]
func (h *Handlers) CreateExecution(r *http.Request) (any, error) {
	var exec mfeatureplat.Execution
	if err := DecodeBody(r, &exec); err != nil {
		return nil, err
	}
	created, err := h.deps.FeatureExecutionSvc.CreateFeatureExecution(&exec)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// CancelExecution godoc
// @Summary      Cancel Execution
// @Description  Cancel a running execution by ID
// @Tags         Executions
// @Param executionId path string true "Execution ID"
// @Produce      json
// @Success      200  {object}  map[string]string "status: cancelled"
// @Security 	 BearerAuth
// @Router       /feature_executions/{executionId}/cancel [put]
func (h *Handlers) CancelExecution(r *http.Request) (any, error) {
	executionId, err := extractExecutionID(r)
	if err != nil {
		return nil, err
	}
	_, err = h.deps.FeatureExecutionSvc.CancelFeatureExecution(executionId)
	if err != nil {
		return nil, err
	}
	return map[string]string{"status": "cancelled"}, nil
}
