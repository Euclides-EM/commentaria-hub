package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
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
// @Description  Create a new dataset. Use async=true to return immediately with status "creating"; poll GET /datasets/{id} for status "ready" or "failed".
// @Tags         Datasets
// @Param		 enforce_single_dataset query bool false "If true, dataset will only be created if no other dataset exists"
// @Param 		 async query bool false "If true, return immediately and create in background (status creating → ready or failed)"
// @Param        dataset  body      model.Dataset  true  "Dataset to create"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   model.Dataset
// @Router       /datasets [post]
func (h *Handlers) CreateDataset(r *http.Request) (any, error) {
	var d model.Dataset
	if err := DecodeBody(r, &d); err != nil {
		return nil, err
	}
	enforceSingleDS, err := strconv.ParseBool(r.URL.Query().Get("enforce_single_dataset"))
	if err != nil {
		enforceSingleDS = false
	}
	async, err := strconv.ParseBool(r.URL.Query().Get("async"))
	if err != nil {
		async = false
	}
	return h.deps.DatasetSvc.Create(r.Context(), &d, enforceSingleDS, async)
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
	if err := DecodeBody(request, &d); err != nil {
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

// UploadDatasetImage godoc
// @Summary      Upload Edition Image
// @Description  Upload an image for a specific edition identified by key. The image file is provided as multipart form data.
// @Tags         Editions
// @Accept       multipart/form-data
// @Produce      json
// @Param        dataSetId  path      string  true  "Dataset ID"
// @Param        key     formData  string  true  "Edition key"
// @Param        type    formData  string  true  "Type of image (e.g., 'cover', 'facsimile')"
// @Param        file    formData  file    true  "Image file to upload"
// @Security 	 BearerAuth
// @Success      200  {object}  model.ImageUpload
// @Router       /datasets/{dataSetId}/images/upload [post]
func (h *Handlers) UploadDatasetImage(r *http.Request) (any, error) {
	datasetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	key := r.FormValue("key")
	if key == "" {
		return nil, fmt.Errorf("key is required for image upload")
	}
	typ := r.FormValue("type")
	if typ == "" {
		return nil, fmt.Errorf("type is required for image upload")
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return h.deps.DatasetSvc.UploadImage(file, header, datasetId, typ, key)
}

// GetDatasetImages godoc
// @Summary      Get Dataset Images
// @Description  Get a list of images associated with a dataset.
// @Tags         Datasets
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Produce      json
// @Success      200  {array}   model.ImageMetadata
// @Router       /datasets/{dataSetId}/images [get]
func (h *Handlers) GetDatasetImages(r *http.Request) (any, error) {
	datasetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	return h.deps.DatasetSvc.ListImages(datasetId)
}
