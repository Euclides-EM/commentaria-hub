package model

type AnnotationUploadMetadata struct {
	DatasetID          string           `json:"dataset_id"`
	Format             AnnotationFormat `json:"format"`
	Name               string           `json:"name"`
	Description        string           `json:"description"`
	Segmented          bool             `json:"segmented"`
	GroundTruth        bool             `json:"ground_truth"`
	Ocred              bool             `json:"ocred"`
	OriginAnnotationID string           `json:"origin_annotation_id,omitempty"`
}
