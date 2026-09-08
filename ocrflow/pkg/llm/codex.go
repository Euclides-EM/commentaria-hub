package llm

import (
	"bufio"
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

const defaultCodexExecutable = "codex"

type CodexClient struct {
	executable string
}

type codexEvent struct {
	Type string `json:"type"`
	Item struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"item"`
	Usage struct {
		InputTokens       int64 `json:"input_tokens"`
		CachedInputTokens int64 `json:"cached_input_tokens"`
		OutputTokens      int64 `json:"output_tokens"`
	} `json:"usage"`
}

func NewCodexClient(executable string) *CodexClient {
	executable = strings.TrimSpace(executable)
	if executable == "" {
		executable = defaultCodexExecutable
	}
	return &CodexClient{executable: executable}
}

func (c *CodexClient) IsAvailable() bool {
	_, err := exec.LookPath(c.executable)
	return err == nil
}

func (c *CodexClient) Exec(model, prompt, attachmentPath string) (string, error) {
	return c.ExecWithLogLabel(model, prompt, attachmentPath, "")
}

func (c *CodexClient) ExecWithLogLabel(model, prompt, attachmentPath, logLabel string) (string, error) {
	result, err := c.ExecResultWithLogLabel(model, prompt, attachmentPath, logLabel)
	return result.Text, err
}

func (c *CodexClient) ExecResultWithLogLabel(model, prompt, attachmentPath, logLabel string) (Result, error) {
	return c.ExecPromptResultWithLogLabel(model, Prompt{Dynamic: prompt}, attachmentPath, logLabel)
}

func (c *CodexClient) ExecPromptResultWithLogLabel(model string, prompt Prompt, attachmentPath, logLabel string) (Result, error) {
	args := []string{"exec", "--ephemeral", "--skip-git-repo-check", "--sandbox", "read-only", "--json", "--model", strings.TrimSpace(model)}
	commandDir := ""
	if strings.TrimSpace(attachmentPath) != "" {
		absolutePath, err := regularAbsolutePath(attachmentPath)
		if err != nil {
			return Result{}, err
		}
		args = append(args, "--image", absolutePath)
		commandDir = filepath.Dir(absolutePath)
	}
	args = append(args, "-")
	return c.run(model, prompt, commandDir, args, logLabel)
}

func (c *CodexClient) ExecWorkspaceResultWithLogLabel(model string, prompt Prompt, workingDir string, _ []string, logLabel string) (Result, error) {
	absoluteDir, err := directoryAbsolutePath(workingDir)
	if err != nil {
		return Result{}, err
	}
	args := []string{"exec", "--ephemeral", "--skip-git-repo-check", "--sandbox", "workspace-write", "--json", "--model", strings.TrimSpace(model), "--cd", absoluteDir, "-"}
	return c.run(model, prompt, absoluteDir, args, logLabel)
}

func (c *CodexClient) run(model string, prompt Prompt, commandDir string, args []string, logLabel string) (Result, error) {
	model = strings.TrimSpace(model)
	combinedPrompt := combinePrompt(prompt)
	if model == "" {
		return Result{}, fmt.Errorf("llm exec: codex model is empty")
	}
	if combinedPrompt == "" {
		return Result{}, fmt.Errorf("llm exec: prompt is empty")
	}
	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	logPrefix := logPrefix(logLabel)
	log.Printf("debug:%s llm exec start provider=codex model=%s workspace=%t", logPrefix, model, commandDir != "")

	cmd := exec.CommandContext(ctx, c.executable, args...)
	cmd.Dir = commandDir
	cmd.Stdin = strings.NewReader(combinedPrompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return Result{}, fmt.Errorf("llm exec: codex timed out after %s: %w", time.Since(startedAt), ctx.Err())
		}
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		return Result{}, fmt.Errorf("llm exec: codex failed after %s: %w: %s", time.Since(startedAt), err, errorBody([]byte(detail)))
	}
	result, err := parseCodexEvents(stdout.Bytes())
	if err != nil {
		return Result{}, err
	}
	log.Printf("debug:%s llm exec end provider=codex model=%s duration=%s error=false tokens_input=%d tokens_cached=%d tokens_output=%d tokens_total=%d", logPrefix, model, time.Since(startedAt), result.Usage.InputTokens, result.Usage.CachedInputTokens, result.Usage.OutputTokens, result.Usage.TotalTokens)
	return result, nil
}

func parseCodexEvents(data []byte) (Result, error) {
	var result Result
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		var event codexEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return Result{}, fmt.Errorf("llm exec: decode codex event: %w: %s", err, errorBody(scanner.Bytes()))
		}
		if event.Type == "item.completed" && event.Item.Type == "agent_message" {
			result.Text = event.Item.Text
		}
		if event.Type == "turn.completed" {
			result.Usage.InputTokens = event.Usage.InputTokens
			result.Usage.CachedInputTokens = event.Usage.CachedInputTokens
			result.Usage.CacheMetricsAvailable = true
			result.Usage.OutputTokens = event.Usage.OutputTokens
			result.Usage.TotalTokens = event.Usage.InputTokens + event.Usage.OutputTokens
		}
	}
	if err := scanner.Err(); err != nil {
		return Result{}, fmt.Errorf("llm exec: read codex events: %w", err)
	}
	if strings.TrimSpace(result.Text) == "" {
		return Result{}, fmt.Errorf("llm exec: codex returned no final message")
	}
	return result, nil
}

func combinePrompt(prompt Prompt) string {
	parts := make([]string, 0, 2)
	if value := strings.TrimSpace(prompt.Static); value != "" {
		parts = append(parts, value)
	}
	if value := strings.TrimSpace(prompt.Dynamic); value != "" {
		parts = append(parts, value)
	}
	return strings.Join(parts, "\n\n")
}

func regularAbsolutePath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("llm exec: resolve attachment %s: %w", path, err)
	}
	info, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("llm exec: access attachment %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("llm exec: attachment %s is not a regular file", path)
	}
	return absolutePath, nil
}

func directoryAbsolutePath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("llm exec: resolve workspace %s: %w", path, err)
	}
	if err := os.MkdirAll(absolutePath, 0o755); err != nil {
		return "", fmt.Errorf("llm exec: create workspace %s: %w", path, err)
	}
	info, err := os.Stat(absolutePath)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("llm exec: workspace %s is not a directory", path)
	}
	return absolutePath, nil
}
