package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestOllamaExecStreamsAndCombinesResponse(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		require.Equal(t, "/api/generate", r.URL.Path)
		require.Equal(t, "application/x-ndjson", r.Header.Get("Accept"))

		var request OllamaGenerateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.True(t, request.Stream)
		require.Equal(t, "vision-model", request.Model)
		require.Equal(t, "correct this", request.Prompt)
		require.Empty(t, request.System)

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/x-ndjson"}},
			Body: io.NopCloser(strings.NewReader(
				"{\"response\":\"corrected \"}\n" +
					"{\"response\":\"text\",\"done\":true}\n",
			)),
			Request: r,
		}, nil
	})

	client := NewOllamaClient("http://ollama.test", "")
	client.httpClient.Transport = transport
	response, err := client.Exec("vision-model", "correct this", "")

	require.NoError(t, err)
	require.Equal(t, "corrected text", response)
}

func TestOllamaExecUsesCompactContractAsSystemPromptAndReportsUsage(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var request OllamaGenerateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, "stable dialect", request.System)
		require.Equal(t, "page input", request.Prompt)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/x-ndjson"}},
			Body: io.NopCloser(strings.NewReader(
				"{\"response\":\"corrected\",\"done\":true,\"prompt_eval_count\":123,\"eval_count\":9}\n",
			)),
			Request: r,
		}, nil
	})
	client := NewOllamaClient("http://ollama.test", "")
	client.httpClient.Transport = transport

	result, err := client.ExecPromptResultWithLogLabel("vision-model", Prompt{
		Static: "stable dialect", Dynamic: "page input",
	}, "", "test")

	require.NoError(t, err)
	require.Equal(t, "corrected", result.Text)
	require.EqualValues(t, 123, result.Usage.InputTokens)
	require.EqualValues(t, 9, result.Usage.OutputTokens)
	require.EqualValues(t, 132, result.Usage.TotalTokens)
	require.False(t, result.Usage.CacheMetricsAvailable)
}

func TestDecodeOllamaGenerateStreamReturnsAPIError(t *testing.T) {
	_, err := decodeOllamaGenerateStream(strings.NewReader("{\"error\":\"model failed\",\"done\":true}\n"))

	require.EqualError(t, err, "llm exec: ollama generate api error: model failed")
}

func TestDecodeOllamaGenerateStreamRejectsIncompleteResponse(t *testing.T) {
	_, err := decodeOllamaGenerateStream(strings.NewReader("{\"response\":\"partial\"}\n"))

	require.EqualError(t, err, "llm exec: ollama generate stream ended before completion")
}
