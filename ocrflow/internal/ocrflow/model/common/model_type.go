package common

type OCRModelType string

const (
	OCRModelTypeSegment OCRModelType = "segment"
	OCRModelTypeOCR     OCRModelType = "text"
	OCRModelTypeUnknown OCRModelType = "unknown"
)

func OCRModelTypeFromExt(ext string) OCRModelType {
	switch ext {
	case ".pt":
		return OCRModelTypeSegment
	case ".mlmodel":
		return OCRModelTypeOCR
	default:
		return OCRModelTypeUnknown
	}
}
