package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeCodeExec(t *testing.T) {
	executable := fakeClaudeCodeExecutable(t, `{"result":"fable response","is_error":false,"subtype":"success"}`)
	client := NewClaudeCodeClient(executable)

	result, err := client.Exec("fable", "Answer briefly", "")
	if err != nil {
		t.Fatal(err)
	}
	if result != "fable response" {
		t.Fatalf("result = %q, want fable response", result)
	}
}

func TestClaudeCodeExecMakesAttachmentAvailableToRead(t *testing.T) {
	executable := fakeClaudeCodeExecutable(t, `{"result":"read attachment","is_error":false,"subtype":"success"}`)
	attachmentPath := filepath.Join(t.TempDir(), "page image.png")
	if err := os.WriteFile(attachmentPath, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
	client := NewClaudeCodeClient(executable)

	if _, err := client.Exec("fable", "Transcribe this page", attachmentPath); err != nil {
		t.Fatal(err)
	}

	invocation, err := os.ReadFile(executable + ".invocation")
	if err != nil {
		t.Fatal(err)
	}
	got := string(invocation)
	if !strings.Contains(got, "--tools\nRead\n") {
		t.Fatalf("invocation did not enable only Read:\n%s", got)
	}
	absolutePath, _ := filepath.Abs(attachmentPath)
	if !strings.Contains(got, absolutePath) {
		t.Fatalf("prompt did not include attachment path %q:\n%s", absolutePath, got)
	}
}

func TestClaudeCodeExecReportsCLIError(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\necho 'authentication required' >&2\nexit 1\n"
	if err := os.WriteFile(executable, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := NewClaudeCodeClient(executable).Exec("fable", "hello", "")
	if err == nil || !strings.Contains(err.Error(), "authentication required") {
		t.Fatalf("error = %v, want authentication detail", err)
	}
}

func fakeClaudeCodeExecutable(t *testing.T, response string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "claude")
	script := `#!/bin/sh
invocation="$0.invocation"
for arg in "$@"; do
  printf '%s\n' "$arg" >> "$invocation"
done
cat >> "$invocation"
printf '%s\n' '` + response + `'
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
