package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/job"
)

// ListIntegrationPlatforms godoc
// @Summary      List Integration Platforms
// @Description  Get a list of supported integration platforms.
// @Tags         Integrations
// @Produce      json
// @Success      200  {array}   integration.Platform
// @Router       /integrations/platforms [get]
func (h *Handlers) ListIntegrationPlatforms(r *http.Request) (any, error) {
	return job.AllTypes, nil
}

// ListIntegrationJobs godoc
// @Summary List integration jobs
// @Description Retrieves a list of all integration jobs
// @Tags Integration
// @Produce json
// @Success 200 {array} integration.Job
// @Failure 500 {string} string "Internal Server Error"
// @Router /integrations/jobs [get]
func (h *Handlers) ListIntegrationJobs(request *http.Request) (any, error) {
	return h.deps.JobSvc.ListIntegrationJobs()
}

// CreateIntegrationJobs godoc
// @Summary Create new integration jobs
// @Description Creates new integration jobs based on the provided details
// @Tags Integration
// @Accept json
// @Produce json
// @Param job body integration.Jobs true "Integration Job Details"
// @Success 201 {object} integration.Jobs
// @Security BearerAuth
// @Router /integrations/jobs [post]
func (h *Handlers) CreateIntegrationJobs(r *http.Request) (any, error) {
	var jobs job.Jobs
	if err := DecodeBody(r, &jobs); err != nil {
		return nil, err
	}
	return h.deps.JobSvc.CreateJobs(&jobs)
}

// GetIntegrationJob godoc
// @Summary Get integration job details
// @Description Retrieves details of a specific integration job by ID
// @Tags Integration
// @Produce json
// @Param jobId path string true "Integration Job ID"
// @Success 200 {object} integration.Job
// @Router /integrations/jobs/{jobId} [get]
func (h *Handlers) GetIntegrationJob(request *http.Request) (any, error) {
	jobId, err := extractJobID(request)
	if err != nil {
		return nil, err
	}
	return h.deps.JobSvc.GetIntegrationJob(jobId)
}
