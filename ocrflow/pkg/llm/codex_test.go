package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodexExecAttachesImageAndReportsUsage(t *testing.T) {
	executable := fakeCodexExecutable(t)
	image := filepath.Join(t.TempDir(), "page image.png")
	require.NoError(t, os.WriteFile(image, []byte("image"), 0o644))

	result, err := NewCodexClient(executable).ExecPromptResultWithLogLabel("gpt-test", Prompt{Static: "contract", Dynamic: "correct page"}, image, "test")
	require.NoError(t, err)
	require.Equal(t, "corrected", result.Text)
	require.EqualValues(t, 120, result.Usage.InputTokens)
	require.EqualValues(t, 20, result.Usage.CachedInputTokens)
	require.EqualValues(t, 30, result.Usage.OutputTokens)
	require.EqualValues(t, 150, result.Usage.TotalTokens)

	invocation, err := os.ReadFile(executable + ".invocation")
	require.NoError(t, err)
	got := string(invocation)
	require.Contains(t, got, "--image\n"+image+"\n")
	require.Contains(t, got, "--sandbox\nread-only\n")
	require.True(t, strings.HasSuffix(got, "contract\n\ncorrect page"))
}

func TestCodexWorkspaceExecutionUsesWritableWorkingDirectory(t *testing.T) {
	executable := fakeCodexExecutable(t)
	workspace := filepath.Join(t.TempDir(), "output")
	_, err := NewCodexClient(executable).ExecWorkspaceResultWithLogLabel("gpt-test", Prompt{Dynamic: "process paths"}, workspace, nil, "test")
	require.NoError(t, err)
	absoluteWorkspace, _ := filepath.Abs(workspace)
	invocation, err := os.ReadFile(executable + ".invocation")
	require.NoError(t, err)
	got := string(invocation)
	require.Contains(t, got, "--sandbox\nworkspace-write\n")
	require.Contains(t, got, "--cd\n"+absoluteWorkspace+"\n")
	require.DirExists(t, workspace)
}

func fakeCodexExecutable(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "codex")
	script := `#!/bin/sh
invocation="$0.invocation"
for arg in "$@"; do
  printf '%s\n' "$arg" >> "$invocation"
done
cat >> "$invocation"
printf '%s\n' '{"type":"item.completed","item":{"type":"agent_message","text":"corrected"}}'
printf '%s\n' '{"type":"turn.completed","usage":{"input_tokens":120,"cached_input_tokens":20,"output_tokens":30}}'
`
	require.NoError(t, os.WriteFile(path, []byte(script), 0o755))
	return path
}
