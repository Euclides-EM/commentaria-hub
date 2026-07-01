package annotation

import (
	"fmt"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type DetectionMode string

const (
	DetectionModeLines        DetectionMode = "lines"
	DetectionModeModelSegment DetectionMode = "model_segment"
	DetectionModeModelOCR     DetectionMode = "model_ocr"
)

func ParseDetectionMode(value string) (DetectionMode, error) {
	switch mode := DetectionMode(strings.TrimSpace(value)); mode {
	case DetectionModeLines, DetectionModeModelSegment, DetectionModeModelOCR:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported detection result mode: %s", value)
	}
}

func DetectionModeForModelType(modelType common.OCRModelType) DetectionMode {
	switch modelType {
	case common.OCRModelTypeOCR:
		return DetectionModeModelOCR
	default:
		return DetectionModeModelSegment
	}
}
