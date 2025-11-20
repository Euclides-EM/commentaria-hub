package httpapi

import (
	"context"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
	_ "github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
	"net/http"
	"time"
)

// Health godoc
// @Summary			Health check
// @Description  	Returns service and DB status
// @Tags         	health
// @Produce      	json
// @Success      	200 {object} models.HealthStatus
// @Router       	/health [get]
func (h *Handlers) Health(r *http.Request) (models.HealthStatus, error) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	return h.deps.HealthService.Check(ctx), nil
}
