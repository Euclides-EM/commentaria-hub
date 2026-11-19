package api

import (
	"fmt"
	"net/http"
	"os"
)

// DownloadJobGPUFarmRunBundle godoc
// @Summary      Download GPU farm run bundle
// @Description  Downloads the bundle of a GPU farm run dispatched by the job with manual_run (annotation detection rules and model training).
// @Tags         Jobs
// @Produce      application/zip
// @Param        jobId path string true "Job ID"
// @Param        runId path string true "GPU farm run ID"
// @Security     BearerAuth
// @Success      200 {file} string "GPU farm run bundle ZIP"
// @Router       /jobs/{jobId}/gpu_farm_runs/{runId}/bundle [get]
func (h *Handlers) DownloadJobGPUFarmRunBundle(r *http.Request) (string, bool, error) {
	bundlePath, err := h.deps.GPUFarmManualBundles.Path(r.PathValue("jobId"), r.PathValue("runId"))
	if err != nil {
		return "", false, err
	}
	if _, err := os.Stat(bundlePath); err != nil {
		return "", false, fmt.Errorf("manual GPU farm bundle not found: %w", err)
	}
	return bundlePath, false, nil
}
