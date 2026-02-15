package model

type Note struct {
	Note string `json:"note"`
}

type ImageUpload struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
}
