package httpapi

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
	"net/http"
	"strings"
)

// ListEditions godoc
// @Summary      List Editions
// @Description  Get a list of available editions. Optionally include facsimiles.
// @Tags         Editions
// @Param        expand  query     string  false  "Include related entities"  Enums(facsimiles)
// @Produce      json
// @Success      200  {array}   models.Edition
// @Router       /editions [get]
func (h *Handlers) ListEditions(r *http.Request) ([]*models.Edition, error) {
	expand := r.URL.Query().Get("expand")
	return h.deps.EditionService.ListEditions(strings.Contains(expand, "facsimiles"))
}

// DownloadFacsimile godoc
// @Summary      Download Facsimile
// @Description  Download a facsimile for a given edition.
// @Tags         Editions
// @Param        editionKey      path      string  true  "Edition Key"
// @Param        id              path      string  true  "Facsimile ID"
// @Param        force_redownload  query     string  false  "Force redownload if already exists"
// @Produce      json
// @Success      200  {object}   map[string]string
// @Router       /editions/{editionKey}/facsimiles/{id}/download [post]
func (h *Handlers) DownloadFacsimile(r *http.Request) error {
	editionKey := r.PathValue("editionKey")
	facsimileID := r.PathValue("id")

	if editionKey == "" || facsimileID == "" {
		return fmt.Errorf("missing parameters")
	}

	forceRedownload := r.URL.Query().Get("force_redownload")

	return h.deps.EditionService.DownloadFacsimile(editionKey, facsimileID, forceRedownload != "")
}
