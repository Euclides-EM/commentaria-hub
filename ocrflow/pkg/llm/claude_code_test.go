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

func TestClaudeCodeExecResultReportsUsageAndCost(t *testing.T) {
	executable := fakeClaudeCodeExecutable(t, `{
		"result":"fable response",
		"is_error":false,
		"subtype":"success",
		"total_cost_usd":0.012345,
		"usage":{
			"input_tokens":100,
			"cache_creation_input_tokens":20,
			"cache_read_input_tokens":30,
			"output_tokens":40
		}
	}`)
	client := NewClaudeCodeClient(executable)

	result, err := client.ExecResultWithLogLabel("fable", "Answer briefly", "", "test")
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "fable response" {
		t.Fatalf("result text = %q, want fable response", result.Text)
	}
	if result.Usage.InputTokens != 100 || result.Usage.CacheCreationInputTokens != 20 ||
		result.Usage.CachedInputTokens != 30 || result.Usage.OutputTokens != 40 || result.Usage.TotalTokens != 190 {
		t.Fatalf("usage = %#v, want decoded Claude Code usage", result.Usage)
	}
	if result.Usage.CostUSD == nil || *result.Usage.CostUSD != 0.012345 {
		t.Fatalf("cost = %v, want 0.012345", result.Usage.CostUSD)
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

func TestClaudeCodeExecSeparatesStaticSystemPromptFromDynamicInput(t *testing.T) {
	executable := fakeClaudeCodeExecutable(t, `{"result":"corrected","is_error":false,"subtype":"success"}`)
	client := NewClaudeCodeClient(executable)

	_, err := client.ExecPromptResultWithLogLabel("fable", Prompt{
		Static:  "stable dialect",
		Dynamic: "page-specific input",
	}, "", "test")
	if err != nil {
		t.Fatal(err)
	}

	invocation, err := os.ReadFile(executable + ".invocation")
	if err != nil {
		t.Fatal(err)
	}
	got := string(invocation)
	if !strings.Contains(got, "--append-system-prompt\nstable dialect\n") {
		t.Fatalf("invocation did not use a stable system prompt:\n%s", got)
	}
	if !strings.HasSuffix(got, "page-specific input") {
		t.Fatalf("stdin did not contain only dynamic input:\n%s", got)
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
