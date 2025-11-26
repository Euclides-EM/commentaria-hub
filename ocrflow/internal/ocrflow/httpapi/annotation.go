package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"net/http"
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
	return h.deps.AnnotationsSvc.ListAnnotations(datasetID)
}

// CreateAnnotation godoc
// @Summary      Create Annotation
// @Description  Create a new annotation for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        async	  	 query     bool    false "Process annotation asynchronously"
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

	return h.deps.AnnotationsSvc.Create(datasetID, &a, r.URL.Query().Get("async") == "true")
}

// ConvertAnnotations godoc
// @Summary      Convert Annotation
// @Description  Convert an existing annotation to a different format.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationConvert  body      model.AnnotationConvert  true  "Annotation conversion details"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/convert [put]
func (h *Handlers) ConvertAnnotations(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")

	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var a model.AnnotationConvert
	if err := decoder.Decode(&a); err != nil {
		return nil, fmt.Errorf("failed to decode annotation convert: %w", err)
	}

	return h.deps.AnnotationsSvc.Convert(datasetID, annotationID, &a)
}

// UploadToRoboflow godoc
// @Summary      Upload Annotation to Roboflow
// @Description  Upload an annotation to Roboflow for a specific dataset.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationRoboflowUpload  body      model.AnnotationRoboflowUpload  true  "Annotation Roboflow upload details"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/rbupload [put]
func (h *Handlers) UploadToRoboflow(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")

	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var urb model.AnnotationRoboflowUpload
	if err := decoder.Decode(&urb); err != nil {
		return nil, fmt.Errorf("failed to decode annotation roboflow upload: %w", err)
	}
	return h.deps.AnnotationsSvc.UploadToRoboflow(datasetID, annotationID, &urb)
}
