package httpapi

import "net/http"

// CleanupLocalStore godoc
// @Summary      Cleanup Local Store
// @Description  Cleans up the local store by removing temporary files and unused data.
// @Tags         Store
// @Param        dry_run  query     string  false  "If true, performs a dry run without deleting files"  Enums(true, false)
// @Success      204  "No Content"
// @Router       /store/cleanup/local [delete]
func (h *Handlers) CleanupLocalStore(r *http.Request) (any, error) {
	return h.deps.MetaStoreManager.CleanupLocalStore(r.URL.Query().Get("dry_run") == "true")
}
