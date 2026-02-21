package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
)

// GetAnnotationTEI godoc
// @Summary      Get Annotation TEI
// @Description  Get the TEI representation of a specific annotation for a specific dataset and page.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id   path      string  true  "Annotation ID"
// @Param        pageNumOrKey   path      string  true  "Page Number or Key"
// @Param        feature   query     []string  false "Features to include in TEI data (can be specified multiple times)"  collectionFormat(multi)
// @Produce      application/xml
// @Success      200  {string}   string "TEI XML content"
// @Router       /datasets/{dataSetId}/annotations/{id}/tei/{pageNumOrKey} [get]
func (h *Handlers) GetAnnotationTEI(r *http.Request) ([]byte, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	pageNumOrKey := r.PathValue("pageNumOrKey")
	if pageNumOrKey == "" {
		return nil, fmt.Errorf("missing page number or key in path")
	}
	features := lo.Map(r.URL.Query()["feature"], func(feature string, _ int) string {
		return strings.TrimSpace(feature)
	})

	return h.deps.AnnotationTEI.GetTEI(datasetID, annotationID, pageNumOrKey, features)
}
