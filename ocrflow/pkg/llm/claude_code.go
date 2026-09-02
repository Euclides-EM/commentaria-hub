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
	Result       string          `json:"result"`
	IsError      bool            `json:"is_error"`
	Subtype      string          `json:"subtype"`
	TotalCostUSD *float64        `json:"total_cost_usd"`
	Usage        claudeCodeUsage `json:"usage"`
}

type claudeCodeUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
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
	result, err := c.ExecResultWithLogLabel(model, prompt, attachmentPath, logLabel)
	return result.Text, err
}

func (c *ClaudeCodeClient) ExecResultWithLogLabel(model, prompt, attachmentPath string, logLabel string) (Result, error) {
	return c.ExecPromptResultWithLogLabel(model, Prompt{Dynamic: prompt}, attachmentPath, logLabel)
}

func (c *ClaudeCodeClient) ExecPromptResultWithLogLabel(model string, prompt Prompt, attachmentPath string, logLabel string) (Result, error) {
	model = strings.TrimSpace(model)
	prompt.Static = strings.TrimSpace(prompt.Static)
	prompt.Dynamic = strings.TrimSpace(prompt.Dynamic)
	if model == "" {
		return Result{}, fmt.Errorf("llm exec: claude code model is empty")
	}
	if prompt.Static == "" && prompt.Dynamic == "" {
		return Result{}, fmt.Errorf("llm exec: prompt is empty")
	}

	args := []string{
		"--print",
		"--output-format", "json",
		"--model", model,
		"--permission-mode", "dontAsk",
		"--no-session-persistence",
		"--tools", "",
	}
	toolsArgIndex := len(args) - 1
	if prompt.Static != "" {
		args = append(args, "--append-system-prompt", prompt.Static)
	}
	dynamicPrompt := prompt.Dynamic
	commandDir := ""
	if strings.TrimSpace(attachmentPath) != "" {
		absolutePath, err := filepath.Abs(attachmentPath)
		if err != nil {
			return Result{}, fmt.Errorf("llm exec: resolve attachment %s: %w", attachmentPath, err)
		}
		info, err := os.Stat(absolutePath)
		if err != nil {
			return Result{}, fmt.Errorf("llm exec: access attachment %s: %w", attachmentPath, err)
		}
		if !info.Mode().IsRegular() {
			return Result{}, fmt.Errorf("llm exec: attachment %s is not a regular file", attachmentPath)
		}

		// Run in the attachment's directory so Read is scoped to the input file's
		// tree without granting Claude Code write or shell tools.
		commandDir = filepath.Dir(absolutePath)
		args[toolsArgIndex] = "Read"
		dynamicPrompt += fmt.Sprintf("\n\nUse the Read tool to inspect the attachment at %q.", absolutePath)
	}

	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	logPrefix := logPrefix(logLabel)
	log.Printf("debug:%s llm exec start provider=claude-code model=%s attachment=%t", logPrefix, model, commandDir != "")

	cmd := exec.CommandContext(ctx, c.executable, args...)
	cmd.Dir = commandDir
	cmd.Stdin = strings.NewReader(dynamicPrompt)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("debug:%s llm exec end provider=claude-code model=%s duration=%s error=true", logPrefix, model, time.Since(startedAt))
		if ctx.Err() != nil {
			return Result{}, fmt.Errorf("llm exec: claude code timed out after %s: %w", time.Since(startedAt), ctx.Err())
		}
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail != "" {
			return Result{}, fmt.Errorf("llm exec: claude code failed after %s: %w: %s", time.Since(startedAt), err, errorBody([]byte(detail)))
		}
		return Result{}, fmt.Errorf("llm exec: claude code failed after %s: %w", time.Since(startedAt), err)
	}

	var response claudeCodeResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return Result{}, fmt.Errorf("llm exec: decode claude code response: %w: %s", err, errorBody(stdout.Bytes()))
	}
	if response.IsError {
		return Result{}, fmt.Errorf("llm exec: claude code returned an error (%s): %s", response.Subtype, strings.TrimSpace(response.Result))
	}

	usage := Usage{
		InputTokens:              response.Usage.InputTokens,
		CachedInputTokens:        response.Usage.CacheReadInputTokens,
		CacheCreationInputTokens: response.Usage.CacheCreationInputTokens,
		CacheMetricsAvailable:    true,
		OutputTokens:             response.Usage.OutputTokens,
		TotalTokens: response.Usage.InputTokens + response.Usage.CacheCreationInputTokens +
			response.Usage.CacheReadInputTokens + response.Usage.OutputTokens,
		CostUSD: response.TotalCostUSD,
	}
	cost := "unavailable"
	if usage.CostUSD != nil {
		cost = fmt.Sprintf("%.6f", *usage.CostUSD)
	}
	log.Printf(
		"debug:%s llm exec end provider=claude-code model=%s duration=%s error=false tokens_input=%d tokens_cached=%d tokens_cache_creation=%d tokens_output=%d tokens_total=%d cost_usd=%s",
		logPrefix,
		model,
		time.Since(startedAt),
		usage.InputTokens,
		usage.CachedInputTokens,
		usage.CacheCreationInputTokens,
		usage.OutputTokens,
		usage.TotalTokens,
		cost,
	)
	return Result{Text: response.Result, Usage: usage}, nil
}
