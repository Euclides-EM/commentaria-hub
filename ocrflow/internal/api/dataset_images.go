package api

import (
	"fmt"
	"net/http"
	"strconv"
)

// UploadDatasetImage godoc
// @Summary      Upload Edition Image
// @Description  Upload an image for a specific edition identified by key. The image file is provided as multipart form data.
// @Tags         Dataset Images
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

	return h.deps.DatasetImgSvc.UploadImage(file, header, datasetId, typ, key)
}

// GetDatasetImages godoc
// @Summary      Get Dataset Images
// @Description  Get a list of images associated with a dataset.
// @Tags         Dataset Images
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        uniqueOnly   query     bool    false  "If true, return only unique images (one per page number or key), otherwise return all images including duplicates for different keys"
// @Produce      json
// @Success      200  {array}   model.ImageMetadata
// @Router       /datasets/{dataSetId}/images [get]
func (h *Handlers) GetDatasetImages(r *http.Request) (any, error) {
	datasetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	uniqueOnlyStr := r.URL.Query().Get("uniqueOnly")
	uniqueOnly := false
	if uniqueOnlyStr != "" {
		var err error
		uniqueOnly, err = strconv.ParseBool(uniqueOnlyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid value for uniqueOnly: %w", err)
		}
	}
	return h.deps.DatasetImgSvc.ListImagesMetadata(datasetId, uniqueOnly)
}

// DeleteDatasetImages godoc
// @Summary      Delete Dataset Image
// @Description  Delete a specific image associated with a dataset.
// @Tags         Dataset Images
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        pageNumOrKey   query 	[]string  false "Page number or image key to identify the image to delete" collectionFormat(multi)
// @Param        filename   query     []string  false  "Filename of the image to delete (optional, used if pageNumOrKey is not sufficient to identify the image)" collectionFormat(multi)
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   map[string]string
// @Router       /datasets/{dataSetId}/images [delete]
func (h *Handlers) DeleteDatasetImages(r *http.Request) (any, error) {
	datasetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	pageNumOrKeys := r.URL.Query()["pageNumOrKey"]
	filenames := r.URL.Query()["filename"]
	if err := h.deps.DatasetImgSvc.DeleteImages(datasetId, pageNumOrKeys, filenames); err != nil {
		return nil, err
	}
	return map[string]string{"status": "deleted"}, nil
}

// GetDatasetImage godoc
// @Summary      Get Page Image
// @Description  Get the image for a specific page in a dataset.
// @Tags         Dataset Images
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        pageNumOrKey   path      string  true  "Page Number"
// @Produce      image/png
// @Success      200  {file}   string "PNG image content"
// @Router       /datasets/{dataSetId}/images/{pageNumOrKey} [get]
func (h *Handlers) GetDatasetImage(r *http.Request) ([]byte, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}

	pageNum := r.PathValue("pageNumOrKey")
	if pageNum == "" {
		return nil, fmt.Errorf("missing page number")
	}
	page, err := strconv.Atoi(pageNum)
	if err != nil {
		return nil, fmt.Errorf("invalid page number: %w", err)
	}
	return h.deps.DatasetImgSvc.GetPageImage(datasetID, page)
}
