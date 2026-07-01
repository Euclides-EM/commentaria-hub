package gpufarm

import (
	"os/exec"
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
