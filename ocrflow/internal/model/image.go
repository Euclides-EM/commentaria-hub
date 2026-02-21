package model

import "time"

type ImageUpload struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

type ImageMetadata struct {
	// Key can be either a page number (as a string) or an image key, depending on how the image was uploaded and identified in the dataset.
	// The key is not necessarily unique across the entire dataset.
	Key string `json:"key"`
	// Filename is the name of the image file, which may not be unique across the dataset. It is unique across the dataset.
	Filename   string    `json:"filename"`
	ModifiedAt time.Time `json:"modified_at"`
}
