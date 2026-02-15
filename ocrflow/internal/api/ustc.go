package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
)

// USTCLookup godoc
// @Summary			Lookup USTC metadata by ID
// @Description  	Fetches metadata from the USTC catalog based on the provided USTC ID.
// @Tags            USTC
// @Param			ustcId path string true "USTC ID to lookup"
// @Produce			json
// @Success			200 {object} model.USTC
// @Security 	    BearerAuth
// @Router          /catalogs/ustc/lookup [post]
func (h *Handlers) USTCLookup(r *http.Request) (any, error) {
	var d model.USTC
	if err := DecodeBody(r, &d); err != nil {
		return nil, err
	}
	return h.deps.USTC.Lookup(d.USTCId)
}
