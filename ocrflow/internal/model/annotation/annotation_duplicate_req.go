package annotation

type DuplicateRequest struct {
	SourceAnnotationID string `json:"source_annotation_id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
}
