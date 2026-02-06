package featureplat

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/common"
	mfeatureplat "github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	"github.com/samber/lo"
)

// ListExecutions godoc
// @Summary      List Executions
// @Description  Get a list of all executions for a specific edition
// @Tags         Executions
// @Produce      json
// @Success      200  {array}   featureplat.FeatureExecution
// @Param collection query string false "Filter by collection ID"
// @Param features query string false "Filter by delimited list of feature IDs"
// @Param statuses query string false "Filter by delimited list of execution statuses" Enums(pending, running, completed, failed)
// @Security 	 BearerAuth
// @Router       /executions [get]
func (h *Handlers) ListExecutions(r *http.Request) (any, error) {
	collection := r.URL.Query().Get("collection")
	features := r.URL.Query().Get("features")
	statuses := r.URL.Query().Get("statuses")

	rawFeatureIds := lo.Map(strings.Split(features, ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})

	var featureIds []string
	if collection != "" || features == "" {
		fs, err := h.deps.FeatureSvc.ListFeatures(nil)
		if err != nil {
			return nil, err
		}
		fsMap := lo.SliceToMap(fs, func(f *mfeatureplat.Feature) (string, *mfeatureplat.Feature) {
			return f.ID, f
		})
		for _, id := range rawFeatureIds {
			if _, ok := fsMap[id]; !ok {
				return nil, fmt.Errorf("feature id %s not found in feature list", id)
			}
			featureIds = append(featureIds, id)
		}
	}

	var featureExecutionsStatuses []mfeatureplat.FeatureExecutionStatus
	for _, s := range strings.Split(statuses, ",") {
		fxs := mfeatureplat.ToFeatureExecutionsStatus(s)
		if fxs == "" {
			return nil, fmt.Errorf("invalid execution status: %s", s)
		}
		featureExecutionsStatuses = append(featureExecutionsStatuses, fxs)
	}

	return h.deps.FeatureExecutionSvc.ListFeatureExecutions(collection, featureIds, featureExecutionsStatuses)
}

// GetExecution godoc
// @Summary      Get Execution
// @Description  Get details of a specific execution by ID
// @Tags         Executions
// @Produce      json
// @Success      200  {object}  featureplat.FeatureExecution
// @Param executionId path string true "Execution ID"
// @Security 	 BearerAuth
// @Router       /executions/{executionId} [get]
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
// @Param        execution  body      featureplat.FeatureExecution  true  "Execution to create"
// @Security 	 BearerAuth
// @Success      200  {object}  featureplat.FeatureExecution
// @Router       /executions [post]
func (h *Handlers) CreateExecution(r *http.Request) (any, error) {
	var exec mfeatureplat.FeatureExecution
	if err := common.DecodeBody(r, &exec); err != nil {
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
// @Router       /executions/{executionId}/cancel [post]
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
