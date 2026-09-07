package service

import (
	"strings"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/gpufarm"
)

func TestDetectionFollowCommandIsDirectlyCopyableFromJSON(t *testing.T) {
	got := detectionFollowCommand(
		"cca-ocr",
		"/pbs/home/m/mjoskowicz/jobs/detect_annotation/run_260831-021856-393/logs/annotation_detect_57325791.out",
		"/pbs/home/m/mjoskowicz/jobs/detect_annotation/run_260831-021856-393/logs/annotation_detect_57325791.err",
	)
	want := "ssh 'cca-ocr' tail -n 100 -F '/pbs/home/m/mjoskowicz/jobs/detect_annotation/run_260831-021856-393/logs/annotation_detect_57325791.out' '/pbs/home/m/mjoskowicz/jobs/detect_annotation/run_260831-021856-393/logs/annotation_detect_57325791.err'"
	if got != want {
		t.Fatalf("detectionFollowCommand() = %q, want %q", got, want)
	}
}

func TestDetectionManifestIncludesFailureCallback(t *testing.T) {
	svc := &AnnotationDetectionRemote{apiURL: "https://example.test/api/v1", apiToken: "secret"}
	ann := &annotation.Annotation{}
	ann.ID = "ann_target"
	ann.DatasetID = "ds_target"
	manifest := svc.detectionManifest(remoteDetectionRequest{
		Mode:       annotation.DetectionModeLines,
		Annotation: ann,
	}, &gpufarm.RemoteEnv{
		RemoteDir:    "/remote/detect",
		RemoteRunDir: "/remote/detect/run_1",
		RunID:        "run_1",
		LogsDir:      "/remote/detect/run_1/logs",
	}, "")
	want := "export RESULT_FAILURE_URL='https://example.test/api/v1/datasets/ds_target/annotations/ann_target/detection_failure'"
	if !strings.Contains(manifest, want) {
		t.Fatalf("manifest missing failure callback URL:\n%s", manifest)
	}
}
