package model

type AnnotationUploadRoboflow struct {
	APIKey           string `json:"api_key"`
	WorkspaceID      string `json:"workspace_url" example:"mia-workplace"`
	ProjectID        string `json:"project_id" example:"paris-1615-xy3bi"`
	IsNotGroundTruth bool   `json:"is_not_ground_truth"`
}
