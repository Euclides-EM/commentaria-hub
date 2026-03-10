package llm

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

const (
	modelGPT5Mini         = "gpt-5-mini"
	unboundedOutputTokens = true
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

	client := openai.NewClient(option.WithAPIKey(c.openAIKey))
	payload := map[string]any{
		"model": modelGPT5Mini,
	}
	if unboundedOutputTokens {
		payload["max_output_tokens"] = nil
	}

	if strings.TrimSpace(attachmentPath) == "" {
		payload["input"] = prompt
	} else {
		fileData, err := os.ReadFile(attachmentPath)
		if err != nil {
			return "", fmt.Errorf("llm exec: read attachment %s: %w", attachmentPath, err)
		}
		mimeType := mime.TypeByExtension(filepath.Ext(attachmentPath))
		if mimeType == "" {
			mimeType = http.DetectContentType(fileData)
		}
		encoded := base64.StdEncoding.EncodeToString(fileData)
		payload["input"] = []map[string]any{
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
		}
	}

	var resp responses.Response
	if err := client.Post(context.Background(), "/responses", payload, &resp); err != nil {
		return "", fmt.Errorf("llm exec: openai responses api call failed: %w", err)
	}
	return resp.OutputText(), nil
}

func NewClient(openAIKey string) *Client {
	return &Client{openAIKey: openAIKey}
}
