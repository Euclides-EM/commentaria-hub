package service

import (
	"context"
	"database/sql"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
)

type Health struct {
	DB *sql.DB
}

func NewHealthService(db *sql.DB) *Health {
	return &Health{DB: db}
}

func (h *Health) Check(ctx context.Context) models.HealthStatus {
	dbOK := false
	if h.DB != nil {
		if err := h.DB.PingContext(ctx); err == nil {
			dbOK = true
		}
	}
	return models.HealthStatus{
		DBReady: dbOK,
	}
}
