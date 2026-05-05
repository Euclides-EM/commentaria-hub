package model

import (
	"fmt"
	"time"
)

type ImageUpload struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Uploaded int    `json:"uploaded"`
}

type ImageMetadata struct {
	// Key can be either a page number (as a string) or an image key, depending on how the image was uploaded and identified in the dataset.
	// The key is not necessarily unique across the entire dataset.
	Key string `json:"key"`
	// Filename is the name of the image file, which may not be unique across the dataset. It is unique across the dataset.
	Filename   string    `json:"filename"`
	ModifiedAt time.Time `json:"modified_at"`
}

type ImageType string

type ImageVariant string

const (
	ImageTypeFacsimile    ImageType = "facsimile"
	ImageTypeTitlePage    ImageType = "tp"
	ImageTypeFrontispiece ImageType = "frontispiece"

	ImageVariantOriginal ImageVariant = "original"
	ImageVariantPreview  ImageVariant = "preview"
	ImageVariantThumb    ImageVariant = "thumb"
)

func ToImageType(s string) (ImageType, error) {
	switch s {
	case "facsimile":
		return ImageTypeFacsimile, nil
	case "tp":
		return ImageTypeTitlePage, nil
	case "frontispiece":
		return ImageTypeFrontispiece, nil
	default:
		return "", fmt.Errorf("invalid image type: %s", s)
	}
}

func ToImageVariant(s string) (ImageVariant, error) {
	switch s {
	case "", "original":
		return ImageVariantOriginal, nil
	case "preview":
		return ImageVariantPreview, nil
	case "thumb":
		return ImageVariantThumb, nil
	default:
		return "", fmt.Errorf("invalid image variant: %s", s)
	}
}
