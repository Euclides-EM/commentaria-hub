package llm

import (
	"fmt"
	"strings"
)

const (
	ProviderOpenAI = "openai"
	ProviderOllama = "ollama"
)

type Client struct {
	openAI *OpenAIClient
	ollama *OllamaClient
}

func NewClient(openAIKey string, ollamaBaseURL string) *Client {
	return &Client{
		openAI: NewOpenAIClient(openAIKey),
		ollama: NewOllamaClient(ollamaBaseURL),
	}
}

func (c *Client) Exec(provider string, model string, prompt string, attachmentPath string) (string, error) {
	var clt AIProviderClient
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case ProviderOpenAI:
		clt = c.openAI
	case ProviderOllama:
		clt = c.ollama
	default:
		return "", fmt.Errorf("llm exec: unsupported ai provider %q", provider)
	}
	return clt.Exec(model, prompt, attachmentPath)
}

type AIProviderClient interface {
	Exec(model string, prompt string, attachmentPath string) (string, error)
}
