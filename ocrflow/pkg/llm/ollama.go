package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
}

type ollamaGenerateRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Stream bool     `json:"stream"`
	Images []string `json:"images,omitempty"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		baseURL: strings.TrimSpace(baseURL),
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (c *OllamaClient) Exec(model, prompt, attachmentPath string) (string, error) {
	if strings.TrimSpace(c.baseURL) == "" {
		return "", fmt.Errorf("llm exec: ollama base url is empty")
	}
	if strings.TrimSpace(model) == "" {
		return "", fmt.Errorf("llm exec: ollama model is empty")
	}
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("llm exec: prompt is empty")
	}

	images, err := buildOllamaImagesPayload(attachmentPath)
	if err != nil {
		return "", err
	}
	payload := ollamaGenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
		Images: images,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("llm exec: encode ollama request: %w", err)
	}

	endpoint, err := ollamaEndpoint(c.baseURL, "/api/generate")
	if err != nil {
		return "", err
	}

	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	log.Printf("debug: llm exec start provider=ollama model=%s attachment=%t", model, strings.TrimSpace(attachmentPath) != "")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("llm exec: create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("debug: llm exec end provider=ollama model=%s duration=%s error=true", model, time.Since(startedAt))
		return "", fmt.Errorf("llm exec: ollama generate api call failed after %s: %w", time.Since(startedAt), err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm exec: read ollama response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		log.Printf("debug: llm exec end provider=ollama model=%s duration=%s status=%d error=true", model, time.Since(startedAt), resp.StatusCode)
		return "", fmt.Errorf("llm exec: ollama generate api returned %d after %s: %s", resp.StatusCode, time.Since(startedAt), truncateForError(respBody))
	}

	var out ollamaGenerateResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("llm exec: decode ollama response: %w", err)
	}
	if strings.TrimSpace(out.Error) != "" {
		log.Printf("debug: llm exec end provider=ollama model=%s duration=%s error=true", model, time.Since(startedAt))
		return "", fmt.Errorf("llm exec: ollama generate api error after %s: %s", time.Since(startedAt), out.Error)
	}

	log.Printf("debug: llm exec end provider=ollama model=%s duration=%s done=%t", model, time.Since(startedAt), out.Done)
	return out.Response, nil
}

func ollamaEndpoint(baseURL string, apiPath string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("llm exec: parse ollama base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("llm exec: ollama base url must include scheme and host")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + apiPath
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func buildOllamaImagesPayload(attachmentPath string) ([]string, error) {
	if strings.TrimSpace(attachmentPath) == "" {
		return nil, nil
	}

	fileData, err := os.ReadFile(attachmentPath)
	if err != nil {
		return nil, fmt.Errorf("llm exec: read attachment %s: %w", attachmentPath, err)
	}
	mimeType := mime.TypeByExtension(filepath.Ext(attachmentPath))
	if mimeType == "" {
		mimeType = http.DetectContentType(fileData)
	}
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("llm exec: ollama attachments must be images, got %s", mimeType)
	}
	return []string{base64.StdEncoding.EncodeToString(fileData)}, nil
}

func truncateForError(body []byte) string {
	const maxLen = 512
	text := strings.TrimSpace(string(body))
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
