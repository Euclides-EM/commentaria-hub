package ocrflow

type AnnotationDuplicateRequest struct {
	SourceAnnotationID string `json:"source_annotation_id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
}
