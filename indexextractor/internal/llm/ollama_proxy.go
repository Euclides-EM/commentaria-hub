package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func ExecOllamaProxy(baseURL string, payload *OllamaGenerateRequest) (*OllamaGenerateResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("llm exec: encode ollama request: %w", err)
	}

	endpoint, err := ollamaEndpoint(baseURL, "/api/generate")
	if err != nil {
		return nil, err
	}

	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("llm exec: create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{Timeout: totalTimeout}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm exec: ollama generate api call failed after %s: %w", time.Since(startedAt), err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("llm exec: read ollama response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("llm exec: ollama generate api returned %d after %s: %s", resp.StatusCode, time.Since(startedAt), truncateForError(respBody))
	}

	var out OllamaGenerateResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("llm exec: decode ollama response: %w", err)
	}
	return &out, nil
}
