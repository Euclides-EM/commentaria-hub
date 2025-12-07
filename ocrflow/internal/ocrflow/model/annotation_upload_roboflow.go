package model

type AnnotationUploadRoboflow struct {
	APIKey           string `json:"api_key"`
	WorkspaceID      string `json:"workspace_url" example:"mia-workplace"`
	ProjectID        string `json:"project_id" example:"dec06miamia-afl6i"`
	IsNotGroundTruth bool   `json:"is_not_ground_truth"`
}
