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

// GetAnnotationTEIs godoc
// @Summary      Get Annotation TEIs
// @Description  Get the TEI representations of all pages for a specific annotation in a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id   path      string  true  "Annotation ID"
// @Param        page   query     string  false "Page numbers to filter TEI data (can be specified multiple times)"
// @Param        key   query     string  false "Page keys to filter TEI data (can be specified multiple times)"
// @Param        feature   query     []string  false "Features to include in TEI data (can be specified multiple times)"  collectionFormat(multi)
// @Produce      application/xml
// @Success      200  {string}   string "TEI XML content for all pages"
// @Router       /datasets/{dataSetId}/annotations/{id}/tei [get]
func (h *Handlers) GetAnnotationTEIs(r *http.Request) ([]byte, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}
	features := lo.Map(r.URL.Query()["feature"], func(feature string, _ int) string {
		return strings.TrimSpace(feature)
	})
	pageNumsOrKeys := append(r.URL.Query()["page"], r.URL.Query()["key"]...)
	if len(pageNumsOrKeys) > 0 {
		pageNumsOrKeys = lo.Map(pageNumsOrKeys, func(p string, _ int) string {
			return strings.TrimSpace(p)
		})
	}

	return h.deps.AnnotationTEI.GetTEIs(datasetID, annotationID, pageNumsOrKeys, features)
}
