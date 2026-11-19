package llm

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

const (
	unboundedOutputTokens = true
	requestTimeout        = 15 * time.Minute
	totalTimeout          = 15 * time.Minute
)

type OpenAIClient struct {
	openAIKey string
}

func (c *OpenAIClient) IsAvailable() bool {
	return strings.TrimSpace(c.openAIKey) != ""
}

func (c *OpenAIClient) Exec(model, prompt, attachmentPath string) (string, error) {
	return c.ExecWithLogLabel(model, prompt, attachmentPath, "")
}

func (c *OpenAIClient) ExecWithLogLabel(model, prompt, attachmentPath string, logLabel string) (string, error) {
	result, err := c.ExecResultWithLogLabel(model, prompt, attachmentPath, logLabel)
	return result.Text, err
}

func (c *OpenAIClient) ExecResultWithLogLabel(model, prompt, attachmentPath string, logLabel string) (Result, error) {
	return c.ExecPromptResultWithLogLabel(model, Prompt{Dynamic: prompt}, attachmentPath, logLabel)
}

func (c *OpenAIClient) ExecPromptResultWithLogLabel(model string, prompt Prompt, attachmentPath string, logLabel string) (Result, error) {
	if strings.TrimSpace(c.openAIKey) == "" {
		return Result{}, fmt.Errorf("llm exec: openai api key is empty")
	}
	if strings.TrimSpace(model) == "" {
		return Result{}, fmt.Errorf("llm exec: openai model is empty")
	}
	if strings.TrimSpace(prompt.Static) == "" && strings.TrimSpace(prompt.Dynamic) == "" {
		return Result{}, fmt.Errorf("llm exec: prompt is empty")
	}

	client := openai.NewClient(
		option.WithAPIKey(c.openAIKey),
		option.WithMaxRetries(0),
	)
	payload := map[string]any{
		"model": model,
	}
	if unboundedOutputTokens {
		payload["max_output_tokens"] = nil
	}

	explicitCaching := strings.TrimSpace(prompt.Static) != "" && supportsExplicitPromptCaching(model)
	if strings.TrimSpace(prompt.CacheKey) != "" {
		payload["prompt_cache_key"] = strings.TrimSpace(prompt.CacheKey)
	}
	if explicitCaching {
		payload["prompt_cache_options"] = map[string]any{"mode": "explicit", "ttl": "30m"}
	}

	input, err := buildInputPayload(prompt, attachmentPath, explicitCaching)
	if err != nil {
		return Result{}, err
	}
	payload["input"] = input

	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	logPrefix := logPrefix(logLabel)
	log.Printf("debug:%s llm exec start provider=openai model=%s attachment=%t", logPrefix, model, strings.TrimSpace(attachmentPath) != "")

	var resp responses.Response
	attempts, err := executeWithRetries(ctx, func() error {
		return client.Post(ctx, "/responses", payload, &resp, option.WithRequestTimeout(requestTimeout))
	})
	if err != nil {
		log.Printf("debug:%s llm exec end provider=openai model=%s duration=%s attempts=%d error=true", logPrefix, model, time.Since(startedAt), attempts)
		return Result{}, fmt.Errorf("llm exec: openai responses api call failed after %s: %w", time.Since(startedAt), err)
	}
	log.Printf(
		"debug:%s llm exec end provider=openai model=%s duration=%s attempts=%d tokens_input=%d tokens_cached=%d tokens_output=%d tokens_reasoning=%d tokens_total=%d",
		logPrefix,
		model,
		time.Since(startedAt),
		attempts,
		resp.Usage.InputTokens,
		resp.Usage.InputTokensDetails.CachedTokens,
		resp.Usage.OutputTokens,
		resp.Usage.OutputTokensDetails.ReasoningTokens,
		resp.Usage.TotalTokens,
	)
	return Result{
		Text: resp.OutputText(),
		Usage: Usage{
			InputTokens:           resp.Usage.InputTokens,
			CachedInputTokens:     resp.Usage.InputTokensDetails.CachedTokens,
			CacheMetricsAvailable: true,
			OutputTokens:          resp.Usage.OutputTokens,
			ReasoningTokens:       resp.Usage.OutputTokensDetails.ReasoningTokens,
			TotalTokens:           resp.Usage.TotalTokens,
		},
	}, nil
}

func NewOpenAIClient(openAIKey string) *OpenAIClient {
	return &OpenAIClient{openAIKey: openAIKey}
}

func supportsExplicitPromptCaching(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	return normalized == "gpt-5.6" || strings.HasPrefix(normalized, "gpt-5.6-")
}

func buildInputPayload(prompt Prompt, attachmentPath string, explicitCaching bool) (any, error) {
	if strings.TrimSpace(attachmentPath) == "" && strings.TrimSpace(prompt.Static) == "" {
		return prompt.Dynamic, nil
	}

	messages := make([]map[string]any, 0, 2)
	if strings.TrimSpace(prompt.Static) != "" {
		staticBlock := map[string]any{
			"type": "input_text",
			"text": prompt.Static,
		}
		if explicitCaching {
			staticBlock["prompt_cache_breakpoint"] = map[string]any{"mode": "explicit"}
		}
		messages = append(messages, map[string]any{
			"role":    "developer",
			"content": []map[string]any{staticBlock},
		})
	}

	dynamicContent := make([]map[string]any, 0, 2)
	if strings.TrimSpace(prompt.Dynamic) != "" {
		dynamicContent = append(dynamicContent, map[string]any{
			"type": "input_text",
			"text": prompt.Dynamic,
		})
	}
	if strings.TrimSpace(attachmentPath) != "" {
		fileData, err := os.ReadFile(attachmentPath)
		if err != nil {
			return nil, fmt.Errorf("llm exec: read attachment %s: %w", attachmentPath, err)
		}
		mimeType := mime.TypeByExtension(filepath.Ext(attachmentPath))
		if mimeType == "" {
			mimeType = http.DetectContentType(fileData)
		}
		encoded := base64.StdEncoding.EncodeToString(fileData)
		dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
		attachment := map[string]any{
			"type":      "input_file",
			"filename":  filepath.Base(attachmentPath),
			"file_data": dataURL,
		}
		if strings.HasPrefix(mimeType, "image/") {
			attachment = map[string]any{
				"type":      "input_image",
				"image_url": dataURL,
			}
		}
		dynamicContent = append(dynamicContent, attachment)
	}
	messages = append(messages, map[string]any{"role": "user", "content": dynamicContent})
	return messages, nil
}
