package service

import (
	"context"
	"database/sql"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
)

type Health struct {
	db *sql.DB
	// todo: add check for github downloader (token valid etc.)
}

func NewHealthService(db *sql.DB) *Health {
	return &Health{db: db}
}

func (h *Health) Check(ctx context.Context) model.HealthStatus {
	dbOK := false
	if h.db != nil {
		if err := h.db.PingContext(ctx); err == nil {
			dbOK = true
		}
	}
	return model.HealthStatus{
		DBReady: dbOK,
	}
}
