package api

import (
	"net/http"
)

// GetEditionDiagramCrops godoc
// @Summary      Get Edition Diagrams
// @Description  Get diagram image URLs for a specific edition key.
// @Tags         Editions
// @Produce      json
// @Param        editionId  path  string  true  "Edition key"
// @Success      200  {object}  model.DiagramCrops
// @Router       /editions/{editionId}/diagrams [get]
func (h *Handlers) GetEditionDiagramCrops(r *http.Request) (any, error) {
	key, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}

	return h.deps.DiagramCropsSvc.GetEditionDiagrams(key)
}
