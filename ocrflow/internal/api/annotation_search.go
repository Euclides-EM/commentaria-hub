package api

import (
	"net/http"
	"strconv"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
)

// SearchAnnotation godoc
// @Summary Search within an annotation's OCR data
// @Description Search for text patterns within specified categories of an annotation's OCR data
// @Tags Annotations
// @Accept json
// @Produce json
// @Param dataSetId path string true "Dataset ID"
// @Param id path string true "Annotation ID"
// @Param category query []string false "Categories to search within (can be specified multiple times)"  collectionFormat(multi)
// @Param search_within query []string false "Fields to search within (categories, transcription, translation, biblio_metadata) (can be specified multiple times)"  collectionFormat(multi)
// @Param regex query string true "Regular expression pattern to search for"
// @Param fallback_to_origin query bool false "Whether to fallback to original annotation"
// @Success 200 {object} annotation.Search
// @Router /datasets/{dataSetId}/annotations/{id}/search [get]
func (h *Handlers) SearchAnnotation(r *http.Request) (any, error) {
	datasetId, annotationId, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}
	categories := r.URL.Query()["category"]
	searchWithin := r.URL.Query()["search_within"]
	regex := r.URL.Query().Get("regex")
	fallbackToOrigin, err := strconv.ParseBool(r.URL.Query().Get("fallback_to_origin"))
	if err != nil {
		fallbackToOrigin = true
	}
	as := &annotation.Search{
		Categories:       categories,
		Regex:            regex,
		DatasetID:        datasetId,
		AnnotationId:     annotationId,
		MaxResults:       500,
		FallbackToOrigin: fallbackToOrigin,
		SearchWithin:     annotation.ToSearchWithin(searchWithin),
	}
	return h.deps.AnnotationSearch.Search(as)
}
