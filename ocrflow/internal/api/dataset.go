package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
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
// @Param        create_default_annotation query bool false "If true, create a default annotation named 'Base' for the dataset"
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
	createDefaultAnnotation, err := strconv.ParseBool(r.URL.Query().Get("create_default_annotation"))
	if err != nil {
		createDefaultAnnotation = false
	}

	var onCreate func(created *model.Dataset) error
	if createDefaultAnnotation {
		onCreate = func(created *model.Dataset) error {
			_, err := h.deps.AnnotationSvc.Create(created.ID, &annotation.Annotation{
				Meta: common.Meta{
					Name: "Base",
				},
			})
			return err
		}
	}

	created, err := h.deps.DatasetSvc.Create(r.Context(), &d, enforceSingleDS, async, onCreate)
	if err != nil {
		return nil, err
	}
	return created, nil
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
