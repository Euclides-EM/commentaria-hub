package api

import (
	"net/http"
	"strconv"
)

// CleanupLocalStore godoc
// @Summary      Cleanup Local Store
// @Description  Cleans up the local store by removing temporary files and unused data.
// @Tags         Store
// @Param        dry_run  query     bool  false  "If true, performs a dry run without deleting files"
// @Security 	 BearerAuth
// @Success      204  "No Content"
// @Router       /store/cleanup/local [delete]
func (h *Handlers) CleanupLocalStore(r *http.Request) (any, error) {
	dryRun, err := strconv.ParseBool(r.URL.Query().Get("dry_run"))
	if err != nil {
		dryRun = false
	}
	ds, err := h.deps.DatasetSvc.CleanupDatasets(dryRun)
	if err != nil {
		return nil, err
	}
	files, err := h.deps.MetaStoreManager.CleanupLocalStore(dryRun)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"files":    files,
		"datasets": ds,
	}, nil
}
