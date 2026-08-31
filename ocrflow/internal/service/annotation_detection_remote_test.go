package service

import "testing"

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
