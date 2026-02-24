package annotation

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type Search struct {
	Categories   []string           `json:"categories"`
	Regex        string             `json:"regex"`
	DatasetID    string             `json:"dataset_id"`
	AnnotationId string             `json:"annotation_id"`
	MaxResults   int                `json:"max_results"`
	Results      []*common.ALTOPart `json:"results" readonly:"true"`
}
