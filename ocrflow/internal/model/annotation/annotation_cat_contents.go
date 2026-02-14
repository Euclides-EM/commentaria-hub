package annotation

type Index struct {
	DatasetID    string       `json:"dataset_id" readonly:"true"`
	AnnotationID string       `json:"annotation_id" readonly:"true"`
	Nodes        []*IndexNode `json:"nodes"`
}

type IndexNode struct {
	Category string       `json:"category"`
	Content  string       `json:"content"`
	Location Location     `json:"location"`
	Children []*IndexNode `json:"children"`
}
