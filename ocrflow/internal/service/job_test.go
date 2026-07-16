package service

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/job"
)

func TestExportJobReadyWithAnnotationReference(t *testing.T) {
	jb := &job.Job{
		Task:       job.Export,
		Target:     &job.Target{Platform: job.PlatformCommentaria},
		Annotation: &annotation.Reference{DatasetID: "dataset-1", ID: "annotation-1"},
	}

	if !isExportJobReady(jb) {
		t.Fatal("export job with annotation reference should be ready")
	}
}

func TestExportJobNotReadyWithoutAnnotationReference(t *testing.T) {
	jb := &job.Job{
		Task:   job.Export,
		Target: &job.Target{Platform: job.PlatformCommentaria},
	}

	if isExportJobReady(jb) {
		t.Fatal("export job without annotation reference should not be ready")
	}
}
