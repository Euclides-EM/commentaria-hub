package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const defaultClaudeCodeExecutable = "claude"

type ClaudeCodeClient struct {
	executable string
}

type claudeCodeResponse struct {
	Result  string `json:"result"`
	IsError bool   `json:"is_error"`
	Subtype string `json:"subtype"`
}

// NewClaudeCodeClient creates a client backed by the locally authenticated
// Claude Code CLI. An empty executable uses "claude" from PATH.
func NewClaudeCodeClient(executable string) *ClaudeCodeClient {
	executable = strings.TrimSpace(executable)
	if executable == "" {
		executable = defaultClaudeCodeExecutable
	}
	return &ClaudeCodeClient{executable: executable}
}

func (c *ClaudeCodeClient) Exec(model, prompt, attachmentPath string) (string, error) {
	return c.ExecWithLogLabel(model, prompt, attachmentPath, "")
}

func (c *ClaudeCodeClient) ExecWithLogLabel(model, prompt, attachmentPath string, logLabel string) (string, error) {
	model = strings.TrimSpace(model)
	prompt = strings.TrimSpace(prompt)
	if model == "" {
		return "", fmt.Errorf("llm exec: claude code model is empty")
	}
	if prompt == "" {
		return "", fmt.Errorf("llm exec: prompt is empty")
	}

	args := []string{
		"--print",
		"--output-format", "json",
		"--model", model,
		"--permission-mode", "dontAsk",
		"--no-session-persistence",
		"--tools", "",
	}
	commandDir := ""
	if strings.TrimSpace(attachmentPath) != "" {
		absolutePath, err := filepath.Abs(attachmentPath)
		if err != nil {
			return "", fmt.Errorf("llm exec: resolve attachment %s: %w", attachmentPath, err)
		}
		info, err := os.Stat(absolutePath)
		if err != nil {
			return "", fmt.Errorf("llm exec: access attachment %s: %w", attachmentPath, err)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("llm exec: attachment %s is not a regular file", attachmentPath)
		}

		// Run in the attachment's directory so Read is scoped to the input file's
		// tree without granting Claude Code write or shell tools.
		commandDir = filepath.Dir(absolutePath)
		args[len(args)-1] = "Read"
		prompt += fmt.Sprintf("\n\nUse the Read tool to inspect the attachment at %q.", absolutePath)
	}

	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	logPrefix := logPrefix(logLabel)
	log.Printf("debug:%s llm exec start provider=claude-code model=%s attachment=%t", logPrefix, model, commandDir != "")

	cmd := exec.CommandContext(ctx, c.executable, args...)
	cmd.Dir = commandDir
	cmd.Stdin = strings.NewReader(prompt)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("debug:%s llm exec end provider=claude-code model=%s duration=%s error=true", logPrefix, model, time.Since(startedAt))
		if ctx.Err() != nil {
			return "", fmt.Errorf("llm exec: claude code timed out after %s: %w", time.Since(startedAt), ctx.Err())
		}
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail != "" {
			return "", fmt.Errorf("llm exec: claude code failed after %s: %w: %s", time.Since(startedAt), err, errorBody([]byte(detail)))
		}
		return "", fmt.Errorf("llm exec: claude code failed after %s: %w", time.Since(startedAt), err)
	}

	var response claudeCodeResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return "", fmt.Errorf("llm exec: decode claude code response: %w: %s", err, errorBody(stdout.Bytes()))
	}
	if response.IsError {
		return "", fmt.Errorf("llm exec: claude code returned an error (%s): %s", response.Subtype, strings.TrimSpace(response.Result))
	}

	log.Printf("debug:%s llm exec end provider=claude-code model=%s duration=%s error=false", logPrefix, model, time.Since(startedAt))
	return response.Result, nil
}
