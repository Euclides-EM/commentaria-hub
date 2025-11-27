package model

import "strings"

type AnnotationFormat string

const (
	AnnotationFormatAlto AnnotationFormat = "alto"
	AnnotationFormatYolo AnnotationFormat = "yolo"
)

func ToAnnotationFormat(format string) AnnotationFormat {
	switch strings.ToLower(format) {
	case "alto":
		return AnnotationFormatAlto
	case "yolo":
		return AnnotationFormatYolo
	default:
		return ""
	}
}
