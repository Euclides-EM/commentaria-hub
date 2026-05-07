package api

import (
	"net/http"
	"strconv"
)

// ListLogs godoc
// @Summary      Tail service logs
// @Description  Returns the last n lines from the deployed API service logs.
// @Tags         Logs
// @Produce      json
// @Param        n  query     int  false  "Number of log lines to return"
// @Success      200 {object} common.LogTail
// @Security     BearerAuth
// @Router       /logs [get]
func (h *Handlers) ListLogs(r *http.Request) (any, error) {
	requestedLines := 0
	if raw := r.URL.Query().Get("n"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return nil, err
		}
		requestedLines = n
	}
	return h.deps.LogsSvc.Tail(r.Context(), requestedLines)
}
