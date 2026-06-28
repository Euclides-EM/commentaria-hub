package api

import "net/http"

// ListPseudonyms godoc
// @Summary      List pseudonyms
// @Description  Returns pseudonyms metadata with name, pseudonym, position, and source.
// @Tags         Metadata
// @Produce      json
// @Success      200  {array}  model.Pseudonym
// @Router       /pseudonyms [get]
func (h *Handlers) ListPseudonyms(_ *http.Request) (any, error) {
	return h.deps.PseudonymSvc.ListPseudonyms()
}
