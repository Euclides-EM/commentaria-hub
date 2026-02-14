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
// @Success      200  {array}   feature.FeatureExecution
// @Param dataset query string false "Filter by dataset ID"
// @Param features query string false "Filter by delimited list of feature IDs"
// @Param statuses query string false "Filter by delimited list of execution statuses" Enums(pending, running, completed, failed)
// @Router       /features/executions [get]
func (h *Handlers) ListExecutions(r *http.Request) (any, error) {
	dataset := r.URL.Query().Get("dataset")
	features := r.URL.Query().Get("features")
	statuses := r.URL.Query().Get("statuses")

	var featureIds []string
	if dataset != "" || features == "" {
		fs, err := h.deps.FeatureSvc.ListFeatures(dataset, nil)
		if err != nil {
			return nil, err
		}
		if dataset != "" {
			fs = lo.Filter(fs, func(f *mfeatureplat.Feature, _ int) bool {
				return f.DatasetID == dataset
			})
		}
		if features == "" {
			for _, f := range fs {
				featureIds = append(featureIds, f.ID)
			}
		}
		if features != "" {
			rawFeatureIds := lo.Map(strings.Split(features, ","), func(s string, _ int) string {
				return strings.TrimSpace(s)
			})
			fs = lo.Filter(fs, func(f *mfeatureplat.Feature, _ int) bool {
				return lo.Contains(rawFeatureIds, f.ID)
			})
		}
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

	return h.deps.FeatureExecutionSvc.ListFeatureExecutions(dataset, featureIds, featureExecutionsStatuses)
}

// GetExecution godoc
// @Summary      Get Execution
// @Description  Get details of a specific execution by ID
// @Tags         Executions
// @Produce      json
// @Success      200  {object}  feature.FeatureExecution
// @Param executionId path string true "Execution ID"
// @Router       /features/executions/{executionId} [get]
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
// @Param        execution  body      feature.FeatureExecution  true  "Execution to create"
// @Security 	 BearerAuth
// @Success      200  {object}  feature.FeatureExecution
// @Router       /features/executions [post]
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
// @Router       /features/executions/{executionId}/cancel [put]
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
