package model

type AnnotationIndex struct {
	DatasetID    string                 `json:"dataset_id" readonly:"true"`
	AnnotationID string                 `json:"annotation_id" readonly:"true"`
	Nodes        []*AnnotationIndexNode `json:"nodes"`
}

type AnnotationIndexNode struct {
	Category string                 `json:"category"`
	Content  string                 `json:"content"`
	Location AnnotationLocation     `json:"location"`
	Children []*AnnotationIndexNode `json:"children"`
}
