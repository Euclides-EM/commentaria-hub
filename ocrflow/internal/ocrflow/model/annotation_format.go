package model

type AnnotationFormat string

const (
	AnnotationFormatAlto AnnotationFormat = "alto"
	AnnotationFormatYolo AnnotationFormat = "yolo"
)

func ToAnnotationFormat(format string) AnnotationFormat {
	switch format {
	case "alto":
		return AnnotationFormatAlto
	case "yolo":
		return AnnotationFormatYolo
	default:
		return ""
	}
}
