package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	"io"
	"net/http"
	"os"
	"strconv"
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
// @Param        random_pages query    number  false "Number of random pages to annotate, only applicable if explicit page list is not provided"
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

	var randomPages int
	rawRandomPages := r.URL.Query().Get("random_pages")
	randomPages, err := strconv.Atoi(rawRandomPages)
	if err != nil && rawRandomPages != "" {
		return nil, fmt.Errorf("invalid random_pages parameter: %w", err)
	}
	return h.deps.AnnotationSvc.Create(datasetID, &a, randomPages)
}

// ApplyRules godoc
// @Summary      Apply Rules to Annotation
// @Description  Apply specific rules to an annotation.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationApplyRules  body 	annotationrule.ApplyRules  true  "Annotation apply rules"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply [put]
func (h *Handlers) ApplyRules(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")

	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var a annotationrule.ApplyRules
	if err := decoder.Decode(&a); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, &a)
}

// ApplyRuleSegment godoc
// @Summary      Apply Segment Rule to Annotation
// @Description  Apply a segment rule to an annotation.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.Segment  true  "Annotation segment rule"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/segment [put]
func (h *Handlers) ApplyRuleSegment(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.Segment
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleSlicePages godoc
// @Summary      Apply Slice Pages Rule to Annotation
// @Description  Apply a slice pages rule to an annotation.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.SlicePages  true  "Annotation slice pages rule"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/slice_pages [put]
func (h *Handlers) ApplyRuleSlicePages(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.SlicePages
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleStretch godoc
// @Summary      Apply Stretch Rule to Annotation
// @Description  Apply a stretch rule to an annotation.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.Stretch  true  "Annotation stretch rule"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/stretch [put]
func (h *Handlers) ApplyRuleStretch(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.Stretch
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleAddMargin godoc
// @Summary      Add Margin Rule to Annotation
// @Description  add margin to an annotation.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.AddMargin  true  "Annotation add margin rule"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/add_margin [put]
func (h *Handlers) ApplyRuleAddMargin(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.AddMargin
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleDetectLines godoc
// @Summary      Detect Lines in Annotation
// @Description  Detect lines in an annotation.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.LinesDetect  true  "Annotation detect lines rule"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/detect_lines [put]
func (h *Handlers) ApplyRuleDetectLines(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.LinesDetect
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleRemoveCategories godoc
// @Summary      Remove Categories in Annotation
// @Description  Remove categories in an annotation.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.RemoveCategories  true  "Remove categories rule"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/remove_categories [put]
func (h *Handlers) ApplyRuleRemoveCategories(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.RemoveCategories
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
}

// ApplyRuleRemoveOverlap godoc
// @Summary      Remove Overlap in Annotation
// @Description  Remove overlapping annotations in an annotation.
// @Tags         Annotations
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        id          path      string  true  "Annotation ID"
// @Param        annotationSegmentRule  body 	annotationrule.RemoveOverlap  true  "Remove overlap rule"
// @Produce      json
// @Success      200  {object}   model.Annotation
// @Router       /datasets/{dataSetId}/annotations/{id}/apply/remove_overlap [put]
func (h *Handlers) ApplyRuleRemoveOverlap(r *http.Request) (any, error) {
	datasetID := r.PathValue("dataSetId")
	annotationID := r.PathValue("id")
	if datasetID == "" || annotationID == "" {
		return nil, fmt.Errorf("missing parameters")
	}

	decoder := json.NewDecoder(r.Body)
	var rule annotationrule.RemoveOverlap
	if err := decoder.Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode annotation apply rules: %w", err)
	}
	rule.Type = rule.GetType()

	rules := &annotationrule.ApplyRules{
		Action: annotationrule.ApplyRulesActionOverwrite,
		Rules: []annotationrule.AnnotationRule{
			&rule,
		},
	}

	return h.deps.AnnotationSvc.ApplyRules(datasetID, annotationID, rules)
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
