package api

import (
	"net/http"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
)

// UploadAnnotationDetectionResult godoc
// @Summary      Upload GPU farm annotation detection result
// @Description  Uploads ALTO result ZIP produced by a GPU farm detection job.
// @Tags         Annotations Apply Rules
// @Accept       multipart/form-data
// @Produce      json
// @Param        dataSetId path string true "Dataset ID"
// @Param        id path string true "Annotation ID"
// @Param        mode formData string true "Detection mode" Enums(lines,model_segment,model_ocr)
// @Param        file formData file true "ALTO result ZIP"
// @Security     BearerAuth
// @Success      200 {object} annotation.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/detection_upload [post]
func (h *Handlers) UploadAnnotationDetectionResult(r *http.Request) (any, error) {
	datasetID, annotationID, err := extractDatasetAndAnnotationIDs(r)
	if err != nil {
		return nil, err
	}
	mode, err := annotation.ParseDetectionMode(r.FormValue("mode"))
	if err != nil {
		return nil, err
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	ann, err := h.deps.AnnotationSvc.UploadDetectionResult(datasetID, annotationID, mode, file)
	if err != nil {
		return nil, err
	}
	if err := h.deps.JobSvc.CompleteAnnotationRuleCallback(datasetID, annotationID); err != nil {
		return nil, err
	}
	return ann, nil
}
