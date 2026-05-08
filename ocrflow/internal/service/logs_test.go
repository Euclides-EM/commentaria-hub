package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestLogsTailUsesNormalizedLineCountAndParsesLines(t *testing.T) {
	svc := NewLogsService("commentaria-hub-api", 200, 500)

	var gotName string
	var gotArgs []string
	svc.run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return []byte("line one\nline two\n"), nil
	}

	resp, err := svc.Tail(context.Background(), 9999)
	if err != nil {
		t.Fatalf("Tail returned error: %v", err)
	}

	if gotName != "journalctl" {
		t.Fatalf("runner name = %q, want %q", gotName, "journalctl")
	}

	wantArgs := []string{"-u", "commentaria-hub-api", "-n", "500", "--no-pager"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("runner args = %#v, want %#v", gotArgs, wantArgs)
	}

	if resp.Service != "commentaria-hub-api" {
		t.Fatalf("response service = %q, want %q", resp.Service, "commentaria-hub-api")
	}
	if resp.Count != 2 {
		t.Fatalf("response count = %d, want %d", resp.Count, 2)
	}
	if !reflect.DeepEqual(resp.Lines, []string{"line one", "line two"}) {
		t.Fatalf("response lines = %#v", resp.Lines)
	}
}

func TestLogsTailReturnsCommandOutputInError(t *testing.T) {
	svc := NewLogsService("commentaria-hub-api", 200, 500)
	svc.run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("permission denied"), errors.New("exit status 1")
	}

	_, err := svc.Tail(context.Background(), 50)
	if err == nil {
		t.Fatal("Tail error = nil, want non-nil")
	}
	if got := err.Error(); got == "" || !containsAll(got, []string{"commentaria-hub-api", "permission denied"}) {
		t.Fatalf("Tail error = %q, expected service name and command output", got)
	}
}

func containsAll(s string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
