package ocrflow

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/common"
	_ "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

// Health godoc
// @Summary			Health check
// @Description  	Returns service and DB status
// @Tags         	Health
// @Produce      	json
// @Success      	200 {object} common.HealthStatus
// @Router       	/health [get]
func (h *Handlers) Health(r *http.Request) (any, error) {
	return common.Health(h.deps.HealthSvc, r)
}
