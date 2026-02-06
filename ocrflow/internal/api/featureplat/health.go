package featureplat

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/common"
)

// Health godoc
// @Summary			Health check for feature app
// @Description  	Returns service and DB status for feature app
// @Tags         	Health
// @Produce      	json
// @Success      	200 {object} common.HealthStatus
// @Router       	/health [get]
func (h *Handlers) Health(r *http.Request) (any, error) {
	return common.Health(h.deps.HealthSvc, r)
}
