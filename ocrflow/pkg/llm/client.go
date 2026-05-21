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
	modelGPT5Mini         = "gpt-5-mini"
	unboundedOutputTokens = true
	requestTimeout        = 5 * time.Minute
	totalTimeout          = 3 * time.Minute
)

type Client struct {
	openAIKey string
}

func (c *Client) Exec(prompt string, attachmentPath string) (string, error) {
	if strings.TrimSpace(c.openAIKey) == "" {
		return "", fmt.Errorf("llm exec: openai api key is empty")
	}
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("llm exec: prompt is empty")
	}

	client := openai.NewClient(
		option.WithAPIKey(c.openAIKey),
		option.WithMaxRetries(0),
	)
	payload := map[string]any{
		"model": modelGPT5Mini,
	}
	if unboundedOutputTokens {
		payload["max_output_tokens"] = nil
	}

	input, err := buildInputPayload(prompt, attachmentPath)
	if err != nil {
		return "", err
	}
	payload["input"] = input

	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	log.Printf("debug: llm exec start model=%s attachment=%t", modelGPT5Mini, strings.TrimSpace(attachmentPath) != "")

	var resp responses.Response
	attempts, err := executeWithRetries(ctx, modelGPT5Mini, func() error {
		return client.Post(ctx, "/responses", payload, &resp, option.WithRequestTimeout(requestTimeout))
	})
	if err != nil {
		log.Printf("debug: llm exec end model=%s duration=%s attempts=%d error=true", modelGPT5Mini, time.Since(startedAt), attempts)
		return "", fmt.Errorf("llm exec: openai responses api call failed after %s: %w", time.Since(startedAt), err)
	}
	log.Printf(
		"debug: llm exec end model=%s duration=%s attempts=%d tokens_input=%d tokens_cached=%d tokens_output=%d tokens_reasoning=%d tokens_total=%d",
		modelGPT5Mini,
		time.Since(startedAt),
		attempts,
		resp.Usage.InputTokens,
		resp.Usage.InputTokensDetails.CachedTokens,
		resp.Usage.OutputTokens,
		resp.Usage.OutputTokensDetails.ReasoningTokens,
		resp.Usage.TotalTokens,
	)
	return resp.OutputText(), nil
}

func NewClient(openAIKey string) *Client {
	return &Client{openAIKey: openAIKey}
}

func buildInputPayload(prompt string, attachmentPath string) (any, error) {
	if strings.TrimSpace(attachmentPath) == "" {
		return prompt, nil
	}

	fileData, err := os.ReadFile(attachmentPath)
	if err != nil {
		return nil, fmt.Errorf("llm exec: read attachment %s: %w", attachmentPath, err)
	}
	mimeType := mime.TypeByExtension(filepath.Ext(attachmentPath))
	if mimeType == "" {
		mimeType = http.DetectContentType(fileData)
	}
	encoded := base64.StdEncoding.EncodeToString(fileData)
	return []map[string]any{
		{
			"role": "user",
			"content": []map[string]any{
				{
					"type":      "input_file",
					"filename":  filepath.Base(attachmentPath),
					"file_data": fmt.Sprintf("data:%s;base64,%s", mimeType, encoded),
				},
				{
					"type": "input_text",
					"text": prompt,
				},
			},
		},
	}, nil
}
