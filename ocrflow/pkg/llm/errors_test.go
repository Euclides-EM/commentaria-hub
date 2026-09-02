package llm

import (
	"strings"
	"testing"
)

func TestErrorBodyKeepsCompleteOutput(t *testing.T) {
	body := "  " + strings.Repeat("diagnostic", 100) + "  "

	got := errorBody([]byte(body))

	if got != strings.TrimSpace(body) {
		t.Fatalf("errorBody() unexpectedly changed provider output")
	}
}
