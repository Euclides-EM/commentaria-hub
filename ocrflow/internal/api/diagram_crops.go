package api

import (
	"io"
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

// GetFacsimileDiagramCrops godoc
// @Summary      Get Facsimile Diagrams
// @Description  Get diagram image URLs for a specific facsimile.
// @Tags         Facsimiles
// @Produce      json
// @Param        id  path  string  true  "Facsimile ID"
// @Success      200  {object}  model.DiagramCrops
// @Router       /facsimilies/{id}/diagrams [get]
func (h *Handlers) GetFacsimileDiagramCrops(r *http.Request) (any, error) {
	id := r.PathValue("id")
	facsimile, err := h.deps.FacsimileSvc.GetFacsimile(id)
	if err != nil {
		return nil, err
	}
	editionFacsimiles, err := h.deps.FacsimileSvc.ListFacsimiles([]string{facsimile.EditionID})
	if err != nil {
		return nil, err
	}
	return h.deps.DiagramCropsSvc.GetFacsimileDiagrams(facsimile, editionFacsimiles)
}

// DownloadEditionDiagramCrops godoc
// @Summary      Download Edition Diagrams
// @Description  Stream a ZIP of the diagram crop images of a specific edition key, laid out as <key>/crops/<image>.
// @Tags         Editions
// @Produce      application/zip
// @Param        editionId  path  string  true  "Edition key"
// @Success      200  {file}  string  "Diagram crops ZIP"
// @Router       /editions/{editionId}/diagrams/zip [get]
func (h *Handlers) DownloadEditionDiagramCrops(r *http.Request) (string, func(io.Writer) error, error) {
	key, err := extractEditionId(r)
	if err != nil {
		return "", nil, err
	}
	return h.deps.DiagramCropsSvc.ZipEditionDiagrams(key)
}

// DownloadFacsimileDiagramCrops godoc
// @Summary      Download Facsimile Diagrams
// @Description  Stream a ZIP of the diagram crop images of a specific facsimile, laid out as <key>/crops/<image>.
// @Tags         Facsimiles
// @Produce      application/zip
// @Param        id  path  string  true  "Facsimile ID"
// @Success      200  {file}  string  "Diagram crops ZIP"
// @Router       /facsimilies/{id}/diagrams/zip [get]
func (h *Handlers) DownloadFacsimileDiagramCrops(r *http.Request) (string, func(io.Writer) error, error) {
	facsimile, err := h.deps.FacsimileSvc.GetFacsimile(r.PathValue("id"))
	if err != nil {
		return "", nil, err
	}
	editionFacsimiles, err := h.deps.FacsimileSvc.ListFacsimiles([]string{facsimile.EditionID})
	if err != nil {
		return "", nil, err
	}
	return h.deps.DiagramCropsSvc.ZipFacsimileDiagrams(facsimile, editionFacsimiles)
}
