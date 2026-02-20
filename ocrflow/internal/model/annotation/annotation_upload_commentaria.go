package annotation

type UploadCommentaria struct {
	APIKey    string `json:"api_key"`
	BasePath  string `json:"base_path" example:"http://euclides.huma-num.fr/commentaria/"`
	DatasetID string `json:"dataset_id"`
}
