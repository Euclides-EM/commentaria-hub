package llm

import (
	"path/filepath"
	"testing"
)

func TestProviderAvailability(t *testing.T) {
	executable := fakeClaudeCodeExecutable(t, `{"result":"ok","is_error":false}`)

	tests := []struct {
		name      string
		available bool
	}{
		{name: "claude code", available: NewClaudeCodeClient(executable).IsAvailable()},
		{name: "missing claude code", available: NewClaudeCodeClient(filepath.Join(t.TempDir(), "missing-claude")).IsAvailable()},
		{name: "openai", available: NewOpenAIClient("key").IsAvailable()},
		{name: "missing openai key", available: NewOpenAIClient(" ").IsAvailable()},
		{name: "ollama", available: NewOllamaClient("https://ollama.example", "").IsAvailable()},
		{name: "missing ollama url", available: NewOllamaClient("", "").IsAvailable()},
	}
	wants := []bool{true, false, true, false, true, false}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.available != wants[i] {
				t.Fatalf("IsAvailable() = %t, want %t", test.available, wants[i])
			}
		})
	}
}

func TestClientIsAvailableRejectsUnknownProvider(t *testing.T) {
	client := NewClient("key", "https://ollama.example", "")
	if client.IsAvailable("unknown") {
		t.Fatal("unknown provider reported as available")
	}
}
