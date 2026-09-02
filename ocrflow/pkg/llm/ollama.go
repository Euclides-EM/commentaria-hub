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

	phttp "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/http"
)

type OllamaClient struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
}

type OllamaGenerateRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	System string   `json:"system,omitempty"`
	Stream bool     `json:"stream"`
	Images []string `json:"images,omitempty"`
}

type OllamaGenerateResponse struct {
	Response        string `json:"response"`
	Done            bool   `json:"done"`
	Error           string `json:"error,omitempty"`
	PromptEvalCount int64  `json:"prompt_eval_count,omitempty"`
	EvalCount       int64  `json:"eval_count,omitempty"`
}

func NewOllamaClient(baseURL, authToken string) *OllamaClient {
	return &OllamaClient{
		baseURL:   strings.TrimSpace(baseURL),
		authToken: strings.TrimSpace(authToken),
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (c *OllamaClient) Exec(model, prompt, attachmentPath string) (string, error) {
	return c.ExecWithLogLabel(model, prompt, attachmentPath, "")
}

func (c *OllamaClient) ExecWithLogLabel(model, prompt, attachmentPath string, logLabel string) (string, error) {
	result, err := c.ExecResultWithLogLabel(model, prompt, attachmentPath, logLabel)
	return result.Text, err
}

func (c *OllamaClient) ExecResultWithLogLabel(model, prompt, attachmentPath string, logLabel string) (Result, error) {
	return c.ExecPromptResultWithLogLabel(model, Prompt{Dynamic: prompt}, attachmentPath, logLabel)
}

func (c *OllamaClient) ExecPromptResultWithLogLabel(model string, prompt Prompt, attachmentPath string, logLabel string) (Result, error) {
	if strings.TrimSpace(c.baseURL) == "" {
		return Result{}, fmt.Errorf("llm exec: ollama base url is empty")
	}
	if strings.TrimSpace(model) == "" {
		return Result{}, fmt.Errorf("llm exec: ollama model is empty")
	}
	if strings.TrimSpace(prompt.Static) == "" && strings.TrimSpace(prompt.Dynamic) == "" {
		return Result{}, fmt.Errorf("llm exec: prompt is empty")
	}

	images, err := buildOllamaImagesPayload(attachmentPath)
	if err != nil {
		return Result{}, err
	}
	payload := OllamaGenerateRequest{
		Model:  model,
		Prompt: prompt.Dynamic,
		System: prompt.Static,
		Stream: true,
		Images: images,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Result{}, fmt.Errorf("llm exec: encode ollama request: %w", err)
	}

	endpoint, err := ollamaEndpoint(c.baseURL, "/api/generate")
	if err != nil {
		return Result{}, err
	}

	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	logPrefix := logPrefix(logLabel)
	log.Printf("debug:%s llm exec start provider=ollama model=%s attachment=%t", logPrefix, model, strings.TrimSpace(attachmentPath) != "")
	var out OllamaGenerateResponse
	attempts, err := executeWithRetriesForAttempts(ctx, maxNetworkRetries, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("llm exec: create ollama request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/x-ndjson")
		if c.authToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.authToken)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("llm exec: read ollama error response: %w", err)
			}
			return phttp.NewHTTPStatusError(resp.StatusCode, resp, fmt.Sprintf("llm exec: ollama generate api returned %d: %s", resp.StatusCode, errorBody(respBody)))
		}

		streamed, err := decodeOllamaGenerateStream(resp.Body)
		if err != nil {
			return err
		}
		out = streamed
		return nil
	})
	if err != nil {
		log.Printf("debug:%s llm exec end provider=ollama model=%s duration=%s attempts=%d error=true", logPrefix, model, time.Since(startedAt), attempts)
		return Result{}, fmt.Errorf("llm exec: ollama generate api call failed after %s: %w", time.Since(startedAt), err)
	}
	log.Printf("debug:%s llm exec end provider=ollama model=%s duration=%s attempts=%d done=%t", logPrefix, model, time.Since(startedAt), attempts, out.Done)
	return Result{
		Text: out.Response,
		Usage: Usage{
			InputTokens:  out.PromptEvalCount,
			OutputTokens: out.EvalCount,
			TotalTokens:  out.PromptEvalCount + out.EvalCount,
		},
	}, nil
}

func decodeOllamaGenerateStream(r io.Reader) (OllamaGenerateResponse, error) {
	decoder := json.NewDecoder(r)
	var response strings.Builder
	var final OllamaGenerateResponse
	chunks := 0

	for {
		var chunk OllamaGenerateResponse
		if err := decoder.Decode(&chunk); err == io.EOF {
			break
		} else if err != nil {
			return OllamaGenerateResponse{}, fmt.Errorf("llm exec: decode ollama stream chunk: %w", err)
		}
		chunks++
		if strings.TrimSpace(chunk.Error) != "" {
			return OllamaGenerateResponse{}, fmt.Errorf("llm exec: ollama generate api error: %s", chunk.Error)
		}
		response.WriteString(chunk.Response)
		if chunk.Done {
			final = chunk
			break
		}
	}

	if chunks == 0 {
		return OllamaGenerateResponse{}, fmt.Errorf("llm exec: ollama generate stream was empty")
	}
	if !final.Done {
		return OllamaGenerateResponse{}, fmt.Errorf("llm exec: ollama generate stream ended before completion")
	}
	final.Response = response.String()
	return final, nil
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
