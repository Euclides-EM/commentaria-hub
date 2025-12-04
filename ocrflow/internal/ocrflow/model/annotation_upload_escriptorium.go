package model

type AnnotationUploadEscriptorium struct {
	BasePath string `json:"base_path"`
	Username string `json:"username"`
	Password string `json:"password"`

	Document string `json:"document" bson:"document"`
}
