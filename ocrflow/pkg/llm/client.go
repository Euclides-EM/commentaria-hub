package llm

import (
	"fmt"
	"strings"
)

const (
	ProviderOpenAI = "openai"
	ProviderOllama = "ollama"
)

var providerConcurrencyLimits = map[string]int{
	ProviderOpenAI: 8,
	ProviderOllama: 1,
}

type Client struct {
	openAI   *OpenAIClient
	ollama   *OllamaClient
	limiters map[string]chan struct{}
}

func NewClient(openAIKey string, ollamaBaseURL, ollamaAuthToken string) *Client {
	return &Client{
		openAI:   NewOpenAIClient(openAIKey),
		ollama:   NewOllamaClient(ollamaBaseURL, ollamaAuthToken),
		limiters: makeProviderLimiters(providerConcurrencyLimits),
	}
}

func (c *Client) Exec(provider string, model string, prompt string, attachmentPath string) (string, error) {
	normalizedProvider := strings.ToLower(strings.TrimSpace(provider))
	var clt AIProviderClient
	switch normalizedProvider {
	case ProviderOpenAI:
		clt = c.openAI
	case ProviderOllama:
		clt = c.ollama
	default:
		return "", fmt.Errorf("llm exec: unsupported ai provider %q", provider)
	}
	slots, ok := c.limiters[normalizedProvider]
	if !ok {
		slots = makeLimiter(1)
		c.limiters[normalizedProvider] = slots
	}
	slots <- struct{}{}
	defer func() {
		<-slots
	}()
	return clt.Exec(model, prompt, attachmentPath)
}

type AIProviderClient interface {
	Exec(model string, prompt string, attachmentPath string) (string, error)
}

func makeLimiter(limit int) chan struct{} {
	if limit < 1 {
		limit = 1
	}
	return make(chan struct{}, limit)
}

func makeProviderLimiters(limits map[string]int) map[string]chan struct{} {
	limiters := make(map[string]chan struct{}, len(limits))
	for provider, limit := range limits {
		limiters[provider] = makeLimiter(limit)
	}
	return limiters
}
