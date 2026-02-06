package common

import (
	"context"
	"net/http"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/service/ocrflow"
)

func Health(healthSvc *ocrflow.Health, r *http.Request) (any, error) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	return healthSvc.Check(ctx), nil
}
