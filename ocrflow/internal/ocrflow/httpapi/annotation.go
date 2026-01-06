package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	"github.com/samber/lo"
)

// ListAnnotations godoc
// @Summary      List Annotations
// @Description  Get a list of annotations for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Produce      json
// @Success      200  {array}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations [get]
func (h *Handlers) ListAnnotations(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	if datasetID == "" {
		return nil, fmt.Errorf("missing dataset ID")
	}
	return h.deps.AnnotationSvc.ListAnnotations(datasetID)
}

// CreateAnnotation godoc
// @Summary      Create Annotation
// @Description  Create a new annotation for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        annotation  body      model.Annotation  true  "Annotation to create"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations [post]
func (h *Handlers) CreateAnnotation(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")

	if datasetID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var a model.Annotation
	if err := decoder.Decode(&a); err != nil {
		return nil, fmt.Errorf("failed to decode annotation: %w", err)
	}

	return h.deps.AnnotationSvc.Create(datasetID, &a)
}

// UploadToRoboflow godoc
// @Summary      Upload Annotation to Roboflow
// @Description  Upload an annotation to Roboflow for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationRoboflowUpload  body      model.AnnotationUploadRoboflow  true  "Annotation Roboflow upload details"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/upload/roboflow [put]
func (h *Handlers) UploadToRoboflow(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")

	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var urb model.AnnotationUploadRoboflow
	if err := decoder.Decode(&urb); err != nil {
		return nil, fmt.Errorf("failed to decode annotation roboflow upload: %w", err)
	}
	return h.deps.AnnotationsUploader.UploadToRoboflow(datasetID, annotationID, &urb)
}

// UploadToEscriptorium godoc
// @Summary      Upload Annotation to Escriptorium
// @Description  Upload an annotation to Escriptorium for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationEscriptoriumUpload  body      model.AnnotationUploadEscriptorium  true  "Annotation Escriptorium upload details"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/upload/escriptorium [put]
func (h *Handlers) UploadToEscriptorium(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")

	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var aue model.AnnotationUploadEscriptorium
	if err := decoder.Decode(&aue); err != nil {
		return nil, fmt.Errorf("failed to decode annotation escriptorium upload: %w", err)
	}
	return h.deps.AnnotationsUploader.UploadToEscriptorium(datasetID, annotationID, &aue)
}

// GetAnnotationZipFile godoc
// @Summary      Upload ZIP File
// @Description  Upload a ZIP file containing annotations.
// @Tags         Annotations
// @Accept       multipart/form-data
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        file  formData  file  true  "ZIP file to upload"
// @Param        format  formData  string  true  "Annotation format (ALTO or YOLO)"
// @Produce      json
// @Success      201  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/fromzip [post]
func (h *Handlers) GetAnnotationZipFile(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	if datasetID == "" {
		return nil, fmt.Errorf("missing dataset ID")
	}
	format := model.AnnotationFormat(r.FormValue("format"))
	if format != model.AnnotationFormatAlto && format != model.AnnotationFormatYolo {
		return nil, fmt.Errorf("unsupported annotation format: %s", format)
	}
	return h.deps.AnnotationSvc.CreateFromZip(datasetID, format, func(dstPath string) error { return httpwrapper.StoreUncompressedDir(dstPath, r) })
}

// GetAnnotationURL godoc
// @Summary      Upload from URL
// @Description  Upload annotations from a ZIP file located at a URL.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        format     query     string  true  "Annotation format (ALTO or YOLO)"
// @Param        url        query     string  true  "URL of the ZIP file to download"
// @Produce      json
// @Success      201  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/fromurl [post]
func (h *Handlers) GetAnnotationURL(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	if datasetID == "" {
		return nil, fmt.Errorf("missing dataset ID")
	}
	format := model.AnnotationFormat(r.FormValue("format"))
	if format != model.AnnotationFormatAlto && format != model.AnnotationFormatYolo {
		return nil, fmt.Errorf("unsupported annotation format: %s", format)
	}
	downloadZipURL := r.FormValue("url")
	if downloadZipURL == "" {
		return nil, fmt.Errorf("missing URL")
	}

	return h.deps.AnnotationSvc.CreateFromZip(datasetID, format, func(dstPath string) error {
		resp, err := http.Get(downloadZipURL)
		if err != nil {
			return fmt.Errorf("failed to download zip from %s: %w", downloadZipURL, err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("failed to download zip from %s: status %s", downloadZipURL, resp.Status)
		}
		src := resp.Body
		defer src.Close()
		dst, err := os.CreateTemp("", "upload-*.zip")
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer dst.Close()
		defer os.Remove(dst.Name())

		_, err = io.Copy(dst, src)
		if err != nil {
			return fmt.Errorf("failed to save file: %w", err)
		}

		if err := futils.Unzip(dst.Name(), dstPath); err != nil {
			return fmt.Errorf("failed to unzip file: %w", err)
		}

		return nil
	})
}

// GetAnnotationIndex godoc
// @Summary      Get Annotation Index
// @Description  Get the index of a specific annotation for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        categories  query     string  true  "Categories for the index"
// @Produce      json
// @Success      200  {object}   model.AnnotationIndex
// @Router       /datasets/{dataSetId}/annotations/{id}/index [get]
func (h *Handlers) GetAnnotationIndex(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	if datasetID == "" {
		return nil, fmt.Errorf("missing dataset ID")
	}
	annotationID := r.PathValue("id")
	if annotationID == "" {
		return nil, fmt.Errorf("missing annotation ID")
	}
	categoriesStr := r.FormValue("categories")
	var categories []string
	if categoriesStr != "" {
		categories = lo.Map(strings.Split(strings.TrimSpace(categoriesStr), ","), func(s string, _ int) string { return strings.TrimSpace(s) })
	}
	return h.deps.AnnotationSvc.GetAnnotationIndex(datasetID, annotationID, categories)
}
