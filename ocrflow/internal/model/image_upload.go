package model

type ImageUpload struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
}
