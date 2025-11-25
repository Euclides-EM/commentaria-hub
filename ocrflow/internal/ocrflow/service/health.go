package service

import (
	"context"
	"database/sql"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
)

type Health struct {
	DB *sql.DB
	// todo: add check for github downloader (token valid etc.)
}

func NewHealthService(db *sql.DB) *Health {
	return &Health{DB: db}
}

func (h *Health) Check(ctx context.Context) model.HealthStatus {
	dbOK := false
	if h.DB != nil {
		if err := h.DB.PingContext(ctx); err == nil {
			dbOK = true
		}
	}
	return model.HealthStatus{
		DBReady: dbOK,
	}
}
