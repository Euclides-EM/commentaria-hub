package model

import "time"

type ImageUpload struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

type ImageMetadata struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	ModifiedAt time.Time `json:"modified_at"`
}
