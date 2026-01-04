package model

type AnnotationCategoryContents struct {
	DatasetID      string           `json:"dataset_id" readonly:"true"`
	AnnotationID   string           `json:"annotation_id" readonly:"true"`
	Category       string           `json:"category" readonly:"true"`
	ContentsByPage map[int][]string `json:"contents" readonly:"true"`
}
