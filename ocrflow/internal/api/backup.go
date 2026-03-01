package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
)

// ListBackups godoc
// @Summary      List backups
// @Description  Returns a list of available backups with their metadata.
// @Tags         Backups
// @Produce      json
// @Success      200  {array}  string
// @Router       /backups [get]
func (h *Handlers) ListBackups(r *http.Request) (any, error) {
	return h.deps.BackupSvc.ListBackups()
}

// CreateBackup godoc
// @Summary      Create backup
// @Description  Creates a new backup of the current system state.
// @Tags         Backups
// @Produce      json
// @Success      201  {string}  string  "Backup ID"
// @Security 	 BearerAuth
// @Router       /backups [post]
func (h *Handlers) CreateBackup(r *http.Request) (any, error) {
	return h.deps.BackupSvc.CreateBackup()
}

// RestoreLatestBackup godoc
// @Summary      Restore latest backup
// @Description  Restores the system state from the latest available backup.
// @Tags         Backups
// @Produce      json
// @Param        backupId   path      string  true  "ID of the backup to restore, if the backupId is 'latest', the latest backup will be restored"
// @Success      200  {object}  map[string]string
// @Security 	 BearerAuth
// @Router       /backups/{backupId}/restore [put]
func (h *Handlers) RestoreLatestBackup(r *http.Request) (any, error) {
	backupId := r.PathValue("backupId")
	if err := h.deps.BackupSvc.SetupRestoreBackup(backupId); err != nil {
		return nil, err
	}
	return map[string]string{"message": "restore initiated, the app will restart"}, nil
}

// DownloadBackup godoc
// @Summary      Download backup
// @Description  Downloads the specified backup file.
// @Tags         Backups
// @Produce      application/octet-stream
// @Param        backupId   path      string  true  "ID of the backup to download"
// @Success      200  {file}  string  "Backup file"
// @Security 	 BearerAuth
// @Router       /backups/{backupId} [get]
func (h *Handlers) DownloadBackup(r *http.Request) (zipPath string, deleteAfterServe bool, err error) {
	backupId := r.PathValue("backupId")
	zipPath, err = h.deps.BackupSvc.GetBackupFullPath(backupId)
	if err != nil {
		return "", false, err
	}
	return zipPath, false, nil
}

// CreateBackupFromZip godoc
// @Summary      Create backup from zip
// @Description  Creates a new backup by uploading a zip file containing the backup data.
// @Tags         Backups
// @Accept       multipart/form-data
// @Param        file  formData  file  true  "Zip file containing the backup data"
// @Produce      json
// @Success      201  {string}  string  "Backup ID"
// @Security 	 BearerAuth
// @Router       /backups/fromzip [post]
func (h *Handlers) CreateBackupFromZip(r *http.Request) (any, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return h.deps.BackupSvc.CreateBackupFromZip(file, func(dstPath string) error { return httpwrapper.StoreUncompressedDir(dstPath, r) })
}
