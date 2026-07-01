package service

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
)

func (a *Annotation) UploadDetectionResult(datasetID string, id string, mode annotation.DetectionMode, zipReader io.Reader) (*annotation.Annotation, error) {
	ann, err := a.Get(datasetID, id)
	if err != nil {
		return nil, err
	}

	release, err := a.acquireRuleRun(ann.DatasetID, ann.ID)
	if err != nil {
		return nil, err
	}
	defer release()

	if err := os.RemoveAll(a.fileSysMgt.DatasetAnnotationYoloDir(ann)); err != nil {
		return nil, fmt.Errorf("failed to remove YOLO dir after ALTO modification: %w", err)
	}

	altoDir := a.fileSysMgt.DatasetAnnotationAltoDir(ann)
	if mode == annotation.DetectionModeModelSegment {
		if err := os.RemoveAll(altoDir); err != nil {
			return nil, fmt.Errorf("failed to clean old ALTO dir: %w", err)
		}
	}
	if err := futils.UnzipFromReader(altoDir, zipReader); err != nil {
		return nil, fmt.Errorf("failed to store GPU farm detection result: %w", err)
	}

	switch mode {
	case annotation.DetectionModeLines:
		ann.LinesDetected = true
	case annotation.DetectionModeModelSegment:
		ann.Segmented = true
	case annotation.DetectionModeModelOCR:
		ann.Ocred = true
	}
	ann.UpdatedAt = time.Now()
	if err := a.annotationStore.UpdateAnnotation(ann); err != nil {
		return nil, fmt.Errorf("failed to update annotation after GPU farm detection upload: %w", err)
	}
	return ann, nil
}
