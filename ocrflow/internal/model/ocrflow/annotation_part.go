package ocrflow

type AnnotationPart struct {
	Category string             `json:"category,omitempty"`
	Content  string             `json:"content,omitempty"`
	Location AnnotationLocation `json:"location"`
}

type AnnotationLocation struct {
	Page        int    `json:"page"`
	TextBlockID string `json:"text_block_id,omitempty"`
}
