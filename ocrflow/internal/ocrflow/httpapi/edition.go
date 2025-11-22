package httpapi

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/models"
	"net/http"
)

// ListEditions godoc
// @Summary      List Editions
// @Description  Get a list of available editions. Optionally include facsimiles.
// @Tags         Editions
// @Param        expand  query     string  false  "Include related entities"  Enums(facsimiles)
// @Param        orderBy query     string  false  "Order by field"            Enums(suggested)
// @Produce      json
// @Success      200  {array}   models.Edition
// @Router       /editions [get]
func (h *Handlers) ListEditions(r *http.Request) ([]*models.Edition, error) {
	return h.deps.EditionService.ListEditions(models.ToEditionExpandOptions(r.URL.Query().Get("expand")), models.ToEditionOrderByOptions("orderBy"))
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
