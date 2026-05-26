package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
)

func (h *Handlers) CreateOllamaRequest(r *http.Request) (any, error) {
	var ollama llm.OllamaGenerateRequest
	if err := DecodeBody(r, &ollama); err != nil {
		return nil, err
	}
	return llm.ExecOllamaProxy(h.deps.Env.OllamaBaseURL, &ollama)
}
