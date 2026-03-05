package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
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
// @Success      200  {array}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations [get]
func (h *Handlers) ListAnnotations(r *http.Request) (any, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	return h.deps.AnnotationSvc.ListAnnotations(datasetID)
}

// CreateAnnotation godoc
// @Summary      Create Annotation
// @Description  Create a new annotation for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        annotation  body      annotation.Annotation  true  "Annotation to create"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations [post]
func (h *Handlers) CreateAnnotation(r *http.Request) (any, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}

	var a annotation.Annotation
	if err = DecodeBody(r, &a); err != nil {
		return nil, err
	}

	return h.deps.AnnotationSvc.Create(datasetID, &a)
}

// GetAnnotation godoc
// @Summary      Get Annotation
// @Description  Get a specific annotation for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id} [get]
func (h *Handlers) GetAnnotation(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}
	return h.deps.AnnotationSvc.Get(datasetID, annotationID)
}

// DeleteAnnotation godoc
// @Summary      Delete Annotation
// @Description  Delete a specific annotation for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        deep		 query     bool    false "Whether to perform a deep delete, which removes all associated files"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   map[string]string
// @Router       /datasets/{dataSetId}/annotations/{id} [delete]
func (h *Handlers) DeleteAnnotation(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}
	fsClean, err := strconv.ParseBool(r.FormValue("deep"))
	if err != nil {
		fsClean = false
	}
	if err := h.deps.AnnotationSvc.Delete(datasetID, annotationID, fsClean); err != nil {
		return nil, fmt.Errorf("failed to delete annotation: %w", err)
	}
	return map[string]string{"status": "deleted"}, nil
}

// UpdateAnnotation godoc
// @Summary      Update Annotation
// @Description  Update a specific annotation for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotation  body      annotation.Annotation  true  "Annotation to update"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id} [put]
func (h *Handlers) UpdateAnnotation(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	var a annotation.Annotation
	if err = DecodeBody(r, &a); err != nil {
		return nil, err
	}

	return h.deps.AnnotationSvc.Update(datasetID, annotationID, &a)
}

// DuplicateAnnotation godoc
// @Summary      Duplicate Annotation
// @Description  Duplicate an existing annotation for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        annotationDuplicateRequest  body      annotation.DuplicateRequest  true  "Annotation duplication details"
// @Security 	 BearerAuth
// @Produce      json
// @Success      200  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/duplicate [post]
func (h *Handlers) DuplicateAnnotation(r *http.Request) (any, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}

	var req annotation.DuplicateRequest
	if err = DecodeBody(r, &req); err != nil {
		return nil, err
	}

	return h.deps.AnnotationSvc.Duplicate(datasetID, req.SourceAnnotationID, req.Name, req.Description)
}

// GetAnnotationZipFile godoc
// @Summary      Upload ZIP File
// @Description  Upload a ZIP file containing annotations.
// @Tags         Annotations
// @Accept       multipart/form-data
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        file  formData  file  true  "ZIP file to upload"
// @Param       format                  query  string true  "Annotation format" Enums(ALTO, YOLO)
// @Param       name                    query  string false "Name of the annotation"
// @Param       description             query  string false "Description of the annotation"
// @Param       segmented               query   bool false "Whether the annotations are segmented"
// @Param       ocred                   query   bool false "Whether the annotations are OCRed"
// @Param       lines_detected         query   bool false "Whether line detection was performed"
// @Param       ground_truth            query  bool false "Whether the annotations are ground truth"
// @Param       origin_annotation_id    query  string false "Origin annotation ID to copy applied rules from"
// @Param        ocr_model_id          query     string  false  "Model ID that was used for OCR processing, only relevant if annotations are OCRed"
// @Param        segment_model_id       query     string  false  "Model ID that was used for segmentation, only relevant if annotations are segmented"
// @Security 	 BearerAuth
// @Produce      json
// @Success      201  {object}   annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/fromzip [post]
func (h *Handlers) GetAnnotationZipFile(r *http.Request) (any, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}

	format := annotation.Format(strings.ToLower(strings.TrimSpace(r.FormValue("format"))))
	if format != annotation.FormatAlto && format != annotation.FormatYolo {
		return nil, fmt.Errorf("unsupported annotation format: %s", format)
	}
	segmented, err := strconv.ParseBool(r.FormValue("segmented"))
	if err != nil {
		segmented = false
	}
	linesDetected, err := strconv.ParseBool(r.FormValue("lines_detected"))
	if err != nil {
		linesDetected = false
	}
	ocred, err := strconv.ParseBool(r.FormValue("ocred"))
	if err != nil {
		ocred = false
	}
	groundTruth, err := strconv.ParseBool(r.FormValue("ground_truth"))
	if err != nil {
		groundTruth = false
	}
	aum := &annotation.UploadMetadata{
		DatasetID:          datasetID,
		Format:             format,
		Name:               r.FormValue("name"),
		Description:        r.FormValue("description"),
		Segmented:          segmented,
		Ocred:              ocred,
		LinesDetected:      linesDetected,
		GroundTruth:        groundTruth,
		OriginAnnotationID: r.FormValue("origin_annotation_id"),
		OCRModelID:         r.FormValue("ocr_model_id"),
		SegmentModelID:     r.FormValue("segment_model_id"),
	}
	return h.deps.AnnotationSvc.CreateFromZip(aum, func(dstPath string) error { return httpwrapper.StoreUncompressedDir(dstPath, r) })
}

// GetAnnotationURL godoc
// @Summary     Upload from URL
// @Description Upload annotations from a ZIP file located at a URL.
// @Tags        Annotations
// @Param       dataSetId               path   string true  "Dataset ID"
// @Param       format                  query  string true  "Annotation format" Enums(ALTO, YOLO)
// @Param       url                     query  string true  "URL of the ZIP file to download"
// @Param       name                    query  string false "Name of the annotation"
// @Param       description             query  string false "Description of the annotation"
// @Param       segmented               query  bool false "Whether the annotations are segmented"
// @Param       ocred                   query  bool false "Whether the annotations are OCRed"
// @Param       lines_detected         query  bool false "Whether line detection was performed"
// @Param       ground_truth            query  bool false "Whether the annotations are ground truth"
// @Param       origin_annotation_id    query  string false "Origin annotation ID to copy applied rules from"
// @Param        ocr_model_id          query     string  false  "Model ID that was used for OCR processing, only relevant if annotations are OCRed"
// @Param        segment_model_id       query     string  false  "Model ID that was used for segmentation, only relevant if annotations are segmented"
// @Security 	 BearerAuth
// @Produce     json
// @Success     201 {object} annotation.Annotation
// @Router      /datasets/{dataSetId}/annotations/fromurl [post]
func (h *Handlers) GetAnnotationURL(r *http.Request) (any, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}

	format := annotation.Format(strings.ToLower(strings.TrimSpace(r.FormValue("format"))))
	if format != annotation.FormatAlto && format != annotation.FormatYolo {
		return nil, fmt.Errorf("unsupported annotation format: %s", format)
	}
	downloadZipURL := r.FormValue("url")
	if downloadZipURL == "" {
		return nil, fmt.Errorf("missing URL")
	}

	segmented, err := strconv.ParseBool(r.FormValue("segmented"))
	if err != nil {
		segmented = false
	}
	linesDetected, err := strconv.ParseBool(r.FormValue("lines_detected"))
	if err != nil {
		linesDetected = false
	}
	ocred, err := strconv.ParseBool(r.FormValue("ocred"))
	if err != nil {
		ocred = false
	}
	groundTruth, err := strconv.ParseBool(r.FormValue("ground_truth"))
	if err != nil {
		groundTruth = false
	}
	aum := &annotation.UploadMetadata{
		DatasetID:          datasetID,
		Format:             format,
		Name:               r.FormValue("name"),
		Description:        r.FormValue("description"),
		Segmented:          segmented,
		Ocred:              ocred,
		Hidden:             false,
		LinesDetected:      linesDetected,
		GroundTruth:        groundTruth,
		OriginAnnotationID: r.FormValue("origin_annotation_id"),
		OCRModelID:         r.FormValue("ocr_model_id"),
		SegmentModelID:     r.FormValue("segment_model_id"),
	}

	return h.deps.AnnotationSvc.CreateFromZip(aum, func(dstPath string) error {
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
// @Param        categories  query     string  false  "Categories for the index"
// @Produce      json
// @Success      200  {object}   annotation.Index
// @Router       /datasets/{dataSetId}/annotations/{id}/index [get]
func (h *Handlers) GetAnnotationIndex(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	categoriesStr := r.FormValue("categories")
	var categories []string
	if categoriesStr != "" {
		categories = lo.Map(strings.Split(strings.TrimSpace(categoriesStr), ","), func(s string, _ int) string { return strings.TrimSpace(s) })
	}
	return h.deps.AnnotationSvc.GetAnnotationIndex(datasetID, annotationID, categories)
}

// ListAnnotationCategories godoc
// @Summary      Get Available Categories
// @Description  Get the available categories for a specific annotation in a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Produce      json
// @Success      200  {array}   string
// @Router       /datasets/{dataSetId}/annotations/{id}/categories [get]
func (h *Handlers) ListAnnotationCategories(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}

	return h.deps.AnnotationSvc.GetAvailableCategories(datasetID, annotationID)
}
