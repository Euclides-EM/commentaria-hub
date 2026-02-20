package model

type ImageUpload struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

type ImageMetadata struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
}
