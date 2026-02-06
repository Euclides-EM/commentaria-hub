package ocrflow

type TrainingStatus string

const (
	TrainingStatusRunning   TrainingStatus = "running"
	TrainingStatusCompleted TrainingStatus = "completed"
	TrainingStatusFailed    TrainingStatus = "failed"
)

type Training struct {
	Meta           `json:",inline"`
	OriginModel    *Model                 `json:"origin_model"`
	AnnotationSets []*AnnotationReference `json:"annotation_sets"`
	Status         TrainingStatus         `json:"status"`
}

type AnnotationReference struct {
	ID        string `json:"id"`
	DatasetID string `json:"dataset_id"`
}
