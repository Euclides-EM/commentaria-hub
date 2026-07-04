package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
)

// GetAnnotationTEI godoc
// @Summary      Get Annotation TEI
// @Description  Get the TEI representation of a specific annotation for a specific dataset and page.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id   path      string  true  "Annotation ID"
// @Param        pageNumOrKey   path      string  true  "Page Number or Key"
// @Param        feature   query     []string  false "Features to include in TEI data (can be specified multiple times)"  collectionFormat(multi)
// @Param        fallback_to_origin   query     bool  true "Whether to fallback to results of the origin annotation if no feature results are found. By default, it's true."
// @Param        image_variant   query     string  false "Optional image coordinate variant: original, preview, or thumb."
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
	features := r.URL.Query()["feature"]
	imageVariant, err := model.ToImageVariant(r.URL.Query().Get("image_variant"))
	if err != nil {
		return nil, err
	}

	fallbackToOrigin, err := strconv.ParseBool(r.URL.Query().Get("fallback_to_origin"))
	if err != nil {
		fallbackToOrigin = true
	}
	return h.deps.AnnotationTEI.GetTEI(datasetID, annotationID, pageNumOrKey, features, fallbackToOrigin, imageVariant)
}

// GetAnnotationTEIs godoc
// @Summary      Get Annotation TEIs
// @Description  Get the TEI representations of a specific annotation for a specific dataset and multiple pages.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id   path      string  true  "Annotation ID"
// @Param        pageNumOrKey   query     []string  true  "Page Numbers or Keys (can be specified multiple times)" collectionFormat(multi)
// @Param        feature   query     []string  false "Features to include in TEI data (can be specified multiple times)"  collectionFormat(multi)
// @Param        fallback_to_origin   query     bool  true "Whether to fallback to results of the origin annotation if no feature results are found. By default, it's true."
// @Param        image_variant   query     string  false "Optional image coordinate variant: original, preview, or thumb."
// @Produce      application/xml
// @Success      200 {string}   string "An XML document containing multiple TEI entries, each representing the TEI for a specific page number or key."
// @Router       /datasets/{dataSetId}/annotations/{id}/teis [get]
func (h *Handlers) GetAnnotationTEIs(r *http.Request) ([]byte, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	rawPageNumOrKeys := r.URL.Query()["pageNumOrKey"]
	var pageNumOrKeys []string
	for _, rawPageNumOrKey := range rawPageNumOrKeys {
		r, err := pagesparser.Range(rawPageNumOrKey)
		if err != nil {
			return nil, fmt.Errorf("invalid page number or key: %s", rawPageNumOrKey)
		}
		pageNumOrKeys = append(pageNumOrKeys, r...)
	}

	if len(pageNumOrKeys) == 0 {
		return nil, fmt.Errorf("missing page number or key in query")
	}
	if len(pageNumOrKeys) > 100 {
		return nil, fmt.Errorf("too many page numbers or keys requested; maximum is 100")
	}
	features := r.URL.Query()["feature"]
	imageVariant, err := model.ToImageVariant(r.URL.Query().Get("image_variant"))
	if err != nil {
		return nil, err
	}

	fallbackToOrigin, err := strconv.ParseBool(r.URL.Query().Get("fallback_to_origin"))
	if err != nil {
		fallbackToOrigin = true
	}
	return h.deps.AnnotationTEI.GetTEIs(datasetID, annotationID, pageNumOrKeys, features, fallbackToOrigin, imageVariant)
}

// GetEditionTEI godoc
// @Summary      Get Edition TEI
// @Description  Get the TEI representation of a specific edition for a specific page.
// @Tags         Editions
// @Param        editionId   path      string  true  "Edition ID"
// @Param        pageNum   path      string  true  "Page Number"
// @Produce      application/xml
// @Success      200  {string}   string "TEI XML content"
// @Router       /editions/{editionId}/tei/{pageNum} [get]
func (h *Handlers) GetEditionTEI(r *http.Request) ([]byte, error) {
	editionID, err := extractEditionId(r)
	if err != nil {
		return nil, err
	}

	pageNum := r.PathValue("pageNum")
	if pageNum == "" {
		return nil, fmt.Errorf("missing page number in path")
	}
	p, err := pagesparser.PageNumber(pageNum)
	if err != nil {
		return nil, fmt.Errorf("invalid page number: %w", err)
	}

	return h.deps.EditionTEI.GetTEI(editionID, p)
}
