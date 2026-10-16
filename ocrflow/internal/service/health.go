package service

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"golang.org/x/sys/unix"
)

type Health struct {
	db        *sql.DB
	vscMgt    *VCSMgt
	storePath string
	// todo: add check for github downloader (token valid etc.)
}

func NewHealthService(db *sql.DB, vscMgt *VCSMgt, storePath ...string) *Health {
	path := ""
	if len(storePath) > 0 {
		path = storePath[0]
	}
	return &Health{db: db, vscMgt: vscMgt, storePath: path}
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
	status := common.HealthStatus{
		DBReady:   dbOK,
		CommitSHA: commitSHA,
	}
	if h.storePath != "" {
		status.Filesystem, status.FilesystemError = filesystemStatus(h.storePath)
	}
	return status
}

func filesystemStatus(path string) (*common.FilesystemStatus, string) {
	reportedPath := path
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		reportedPath = resolved
	}
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return nil, fmt.Sprintf("stat filesystem for %s: %v", path, err)
	}
	blockSize := uint64(stat.Bsize)
	total := stat.Blocks * blockSize
	free := stat.Bfree * blockSize
	available := stat.Bavail * blockSize
	used := total - free
	usedPercent := 0.0
	if total > 0 {
		usedPercent = float64(used) * 100 / float64(total)
	}
	return &common.FilesystemStatus{
		Path:           reportedPath,
		TotalBytes:     total,
		UsedBytes:      used,
		AvailableBytes: available,
		UsedPercent:    usedPercent,
	}, ""
}
