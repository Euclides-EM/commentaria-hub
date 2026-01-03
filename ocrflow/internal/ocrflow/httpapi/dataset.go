package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/querylang"
)

// ListDatasets godoc
// @Summary      List Datasets
// @Description  Get a list of datasets with optional filtering and sorting.
// @Tags         Datasets
// @Param        filter  query     string  false  "Filter conditions"
// @Param        sort    query     string  false  "Sort conditions"
// @Produce      json
// @Success      200  {array}   model.Dataset
// @Router       /datasets [get]
func (h *Handlers) ListDatasets(r *http.Request) (any, error) {
	filter, err := querylang.ParseFilter(r.URL.Query().Get("filter"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse filter: %w", err)
	}
	sort, err := querylang.ParseSort(r.URL.Query().Get("sort"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse sort: %w", err)
	}
	return h.deps.DatasetSvc.List(filter, sort)
}

// CreateDataset godoc
// @Summary      Create Dataset
// @Description  Create a new dataset.
// @Tags         Datasets
// @Param        force_overwrite  query     string  false  "Force overwrite if dataset already exists"
// @Param        skip_deskew      query     string  false  "Skip deskewing of images"
// @Param        dataset  body      model.Dataset  true  "Dataset to create"
// @Produce      json
// @Success      200  {object}   model.Dataset
// @Router       /datasets [post]
func (h *Handlers) CreateDataset(r *http.Request) (any, error) {
	decoder := json.NewDecoder(r.Body)
	var d model.Dataset
	if err := decoder.Decode(&d); err != nil {
		return nil, fmt.Errorf("failed to decode request body: %w", err)
	}
	return h.deps.DatasetSvc.Create(&d,
		strings.ToLower(strings.TrimSpace(r.URL.Query().Get("force_overwrite"))) == "true",
		strings.ToLower(strings.TrimSpace(r.URL.Query().Get("skip_deskew"))) == "true",
	)
}

// ListSuggestedRulesForDataset godoc
// @Summary      List Suggested Annotation Rules for Dataset
// @Description  Get a list of suggested annotation rules for a specific dataset.
// @Tags         Datasets
// @Param        dataSetId  path      string  true  "Dataset ID"
// @Produce      json
// @Success      200  {array}   [][]annotationrule.AnnotationRule
// @Router       /datasets/{dataSetId}/suggested_rules [get]
func (h *Handlers) ListSuggestedRulesForDataset(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	return h.deps.DatasetSvc.ListSuggestedAnnotationRules(datasetID)
}

// GetPageImage godoc
// @Summary      Get Page Image
// @Description  Get the image for a specific page in a dataset.
// @Tags         Datasets
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        pageNum   path      string  true  "Page Number"
// @Produce      image/png
// @Success      200  {file}   string "PNG image content"
// @Router       /datasets/{dataSetId}/pages/{pageNum}/image [get]
func (h *Handlers) GetPageImage(r *http.Request) ([]byte, error) {
	datasetID := r.PathValue("dataSetId")
	if datasetID == "" {
		return nil, fmt.Errorf("missing dataset ID")
	}
	pageNum := r.PathValue("pageNum")
	if pageNum == "" {
		return nil, fmt.Errorf("missing page number")
	}
	page, err := strconv.Atoi(pageNum)
	if err != nil {
		return nil, fmt.Errorf("invalid page number: %w", err)
	}
	return h.deps.DatasetSvc.GetPageImage(datasetID, page)
}
