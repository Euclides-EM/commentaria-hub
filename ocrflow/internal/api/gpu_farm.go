package api

import (
	"fmt"
	"net/http"
	"os"
)

// DownloadGPUFarmManualBundle godoc
// @Summary      Download manual GPU farm run bundle
// @Description  Downloads the bundle of a GPU farm run submitted with manual_run (annotation detection rules and model training).
// @Tags         GPU Farm
// @Produce      application/zip
// @Param        runId path string true "GPU farm run ID"
// @Security     BearerAuth
// @Success      200 {file} string "Manual run bundle ZIP"
// @Router       /gpu_farm/manual_bundles/{runId} [get]
func (h *Handlers) DownloadGPUFarmManualBundle(r *http.Request) (string, bool, error) {
	bundlePath, err := h.deps.GPUFarmManualBundles.Path(r.PathValue("runId"))
	if err != nil {
		return "", false, err
	}
	if _, err := os.Stat(bundlePath); err != nil {
		return "", false, fmt.Errorf("manual GPU farm bundle not found: %w", err)
	}
	return bundlePath, false, nil
}
