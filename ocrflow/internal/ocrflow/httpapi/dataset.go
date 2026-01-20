package httpapi

import (
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
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Dataset
// @Router       /datasets [post]
func (h *Handlers) CreateDataset(r *http.Request) (any, error) {
	var d model.Dataset
	if err := decodeBody(r, &d); err != nil {
		return nil, err
	}
	return h.deps.DatasetSvc.Create(r.Context(), &d,
		strings.ToLower(strings.TrimSpace(r.URL.Query().Get("force_overwrite"))) == "true",
		strings.ToLower(strings.TrimSpace(r.URL.Query().Get("skip_deskew"))) == "true",
	)
}

// DeleteDataset godoc
// @Summary      Delete Dataset
// @Description  Delete a dataset by its ID.
// @Tags         Datasets
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   map[string]string
// @Router       /datasets/{dataSetId} [delete]
func (h *Handlers) DeleteDataset(r *http.Request) (any, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	if err = h.deps.DatasetSvc.Delete(datasetID); err != nil {
		return nil, err
	}
	return map[string]string{"status": "deleted"}, nil
}

// UpdateDataset godoc
// @Summary      Update Dataset
// @Description  Update an existing dataset.
// @Tags         Datasets
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        dataset  body      model.Dataset  true  "Updated dataset"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Dataset
// @Router       /datasets/{dataSetId} [put]
func (h *Handlers) UpdateDataset(request *http.Request) (any, error) {
	datasetID, err := extractDatasetID(request)
	if err != nil {
		return nil, err
	}
	var d model.Dataset
	if err := decodeBody(request, &d); err != nil {
		return nil, err
	}
	return h.deps.DatasetSvc.Update(datasetID, &d)
}

// GetPageImage godoc
// @Summary      Get Page Image
// @Description  Get the image for a specific page in a dataset.
// @Tags         Datasets
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        pageNum   path      string  true  "Page Number"
// @Produce      image/png
// @Success      200  {file}   string "PNG image content"
// @Router       /datasets/{dataSetId}/images/{pageNum} [get]
func (h *Handlers) GetPageImage(r *http.Request) ([]byte, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
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
