package annotation

type Search struct {
	Categories   []string `json:"categories"`
	Regex        string   `json:"regex"`
	DatasetID    string   `json:"dataset_id"`
	AnnotationId string   `json:"annotation_id"`
	MaxResults   int      `json:"max_results"`
	Results      []*Part  `json:"results" readonly:"true"`
}
