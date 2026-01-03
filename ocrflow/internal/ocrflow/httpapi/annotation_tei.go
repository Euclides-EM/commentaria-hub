package httpapi

import (
	"fmt"
	"net/http"
)

// GetAnnotationTEI godoc
// @Summary      Get Annotation TEI
// @Description  Get the TEI representation of a specific annotation for a specific dataset and page.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        annotationId   path      string  true  "Annotation ID"
// @Param        pageNum   path      string  true  "Page Number"
// @Produce      application/xml
// @Success      200  {string}   string "TEI XML content"
// @Router       /datasets/{dataSetId}/annotations/{annotationId}/tei/{pageNum} [get]
func (h *Handlers) GetAnnotationTEI(r *http.Request) ([]byte, error) {
	datasetID := r.PathValue("dataSetId")
	if datasetID == "" {
		return nil, fmt.Errorf("missing dataset ID")
	}
	annotationID := r.PathValue("annotationId")
	if annotationID == "" {
		return nil, fmt.Errorf("missing annotation ID")
	}
	pageNum := r.PathValue("pageNum")
	if pageNum == "" {
		return nil, fmt.Errorf("missing page number")
	}

	return h.deps.AnnotationTEI.GetTEI(datasetID, annotationID, pageNum)
}
