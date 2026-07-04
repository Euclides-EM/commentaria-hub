package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

type Health struct {
	db     *sql.DB
	vscMgt *VCSMgt
	// todo: add check for github downloader (token valid etc.)
}

func NewHealthService(db *sql.DB, vscMgt *VCSMgt) *Health {
	return &Health{db: db, vscMgt: vscMgt}
}

func (h *Health) Check(ctx context.Context) common.HealthStatus {
	dbOK := false
	if h.db != nil {
		if err := h.db.PingContext(ctx); err == nil {
			dbOK = true
		}
	}
	commitSHA := ""
	if h.vscMgt != nil {
		if cs, err := h.vscMgt.GetCommitSHA(h.vscMgt.repoPath); err == nil {
			commitSHA = cs
		} else {
			commitSHA = fmt.Sprintf("error: %v", err)
		}
	}
	return common.HealthStatus{
		DBReady:   dbOK,
		CommitSHA: commitSHA,
	}
}
