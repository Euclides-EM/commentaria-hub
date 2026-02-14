package api

import (
	"context"
	"net/http"
	"time"

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
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	return h.deps.HealthSvc.Check(ctx), nil
}
