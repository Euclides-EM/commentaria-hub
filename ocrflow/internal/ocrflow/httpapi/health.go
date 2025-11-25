package httpapi

import (
	"context"
	_ "github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"net/http"
	"time"
)

// Health godoc
// @Summary			Health check
// @Description  	Returns service and DB status
// @Tags         	health
// @Produce      	json
// @Success      	200 {object} model.HealthStatus
// @Router       	/health [get]
func (h *Handlers) Health(r *http.Request) (any, error) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	return h.deps.HealthSvc.Check(ctx), nil
}
