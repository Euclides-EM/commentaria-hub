package gpufarm

import (
	"os/exec"
	"strings"
	"testing"
)

func TestRenderCleanupCompletedRunsScriptIsValidBash(t *testing.T) {
	script, err := renderCleanupCompletedRunsScript("/tmp/jobs with ' quote", 3)
	if err != nil {
		t.Fatalf("render cleanup script: %v", err)
	}

	if output, err := exec.Command("bash", "-n", "-c", script).CombinedOutput(); err != nil {
		t.Fatalf("validate cleanup script: %v: %s", err, output)
	}
}

func TestCleanupScriptReportsRetentionDecisions(t *testing.T) {
	script, err := renderCleanupCompletedRunsScript("/tmp/jobs", 3)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"reason=active",
		"abandoned_upload_older_than_24h",
		"retention_limit",
		"deleted_abandoned=",
		"deleted_retention=",
	} {
		if !strings.Contains(script, expected) {
			t.Fatalf("cleanup script lacks %q", expected)
		}
	}
}

func TestDiscardRejectsPathsOutsideDirectJobRunDirectory(t *testing.T) {
	s := NewSubmitterSlurm("unused", "/remote/jobs")
	for _, runDir := range []string{
		"/remote/jobs",
		"/remote/jobs/detect_annotation",
		"/remote/jobs/detect_annotation/not-a-run",
		"/remote/jobs/detect_annotation/nested/run_123",
		"/remote/elsewhere/detect_annotation/run_123",
	} {
		if err := s.Discard(&RemoteEnv{RemoteRunDir: runDir}); err == nil {
			t.Errorf("Discard(%q) unexpectedly succeeded", runDir)
		}
	}
}
