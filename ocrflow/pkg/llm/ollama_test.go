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

func TestDecodeOllamaGenerateStreamReturnsAPIError(t *testing.T) {
	_, err := decodeOllamaGenerateStream(strings.NewReader("{\"error\":\"model failed\",\"done\":true}\n"))

	require.EqualError(t, err, "llm exec: ollama generate api error: model failed")
}

func TestDecodeOllamaGenerateStreamRejectsIncompleteResponse(t *testing.T) {
	_, err := decodeOllamaGenerateStream(strings.NewReader("{\"response\":\"partial\"}\n"))

	require.EqualError(t, err, "llm exec: ollama generate stream ended before completion")
}
