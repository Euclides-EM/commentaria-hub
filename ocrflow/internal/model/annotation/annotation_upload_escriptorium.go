package annotation

type UploadEscriptorium struct {
	BasePath string `json:"base_path" example:""`
	Username string `json:"username" example:""`
	Password string `json:"password" example:""`

	Document string `json:"document"`
}
