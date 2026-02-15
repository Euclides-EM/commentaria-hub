package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
)

// USTCLookup godoc
// @Summary			Lookup USTC metadata by ID
// @Description  	Fetches metadata from the USTC catalog based on the provided USTC ID.
// @Tags            USTC
// @Param			ustc body model.USTC true "JSON with ustc_id"
// @Produce			json
// @Success			200 {object} model.USTC
// @Security 	    BearerAuth
// @Router          /catalogs/ustc/lookup [post]
func (h *Handlers) USTCLookup(r *http.Request) (any, error) {
	var d model.USTC
	if q := r.URL.Query().Get("ustc_id"); q != "" {
		id, err := strconv.Atoi(q)
		if err != nil {
			return nil, fmt.Errorf("invalid ustc_id query: %w", err)
		}
		d.USTCId = id
	} else {
		if err := DecodeBody(r, &d); err != nil {
			if errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("request body required: send JSON {\"ustc_id\": <number>} or query ?ustc_id=<number>")
			}
			return nil, err
		}
	}
	if d.USTCId <= 0 {
		return nil, fmt.Errorf("ustc_id must be a positive integer")
	}
	return h.deps.USTC.Lookup(d.USTCId)
}
