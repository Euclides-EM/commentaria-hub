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
// @Success      200  {array}   job.Platform
// @Router       /integrations/platforms [get]
func (h *Handlers) ListIntegrationPlatforms(r *http.Request) (any, error) {
	return job.AllTypes, nil
}

// ListJobs godoc
// @Summary List jobs
// @Description Retrieves a list of all jobs
// @Tags Jobs
// @Produce json
// @Success 200 {array} job.Job
// @Failure 500 {string} string "Internal Server Error"
// @Router /jobs [get]
func (h *Handlers) ListJobs(request *http.Request) (any, error) {
	return h.deps.JobSvc.ListJobs()
}

// CreateJobs godoc
// @Summary Create new jobs
// @Description Creates new jobs based on the provided details
// @Tags Jobs
// @Accept json
// @Produce json
// @Param job body job.Jobs true "Job details"
// @Success 201 {object} job.Jobs
// @Security BearerAuth
// @Router /jobs [post]
func (h *Handlers) CreateJobs(r *http.Request) (any, error) {
	var jobs job.Jobs
	if err := DecodeBody(r, &jobs); err != nil {
		return nil, err
	}
	return h.deps.JobSvc.CreateJobs(&jobs)
}

// GetJob godoc
// @Summary Get job details
// @Description Retrieves details of a specific job by ID
// @Tags Jobs
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} job.Job
// @Router /jobs/{jobId} [get]
func (h *Handlers) GetJob(request *http.Request) (any, error) {
	jobId, err := extractJobID(request)
	if err != nil {
		return nil, err
	}
	return h.deps.JobSvc.GetJob(jobId)
}
