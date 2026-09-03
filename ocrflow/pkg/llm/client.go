package llm

import (
	"fmt"
	"strings"
)

const (
	ProviderClaudeCode = "claude-code"
	ProviderOpenAI     = "openai"
	ProviderOllama     = "ollama"
)

var providerConcurrencyLimits = map[string]int{
	ProviderClaudeCode: 4,
	ProviderOpenAI:     8,
	ProviderOllama:     1,
}

type Client struct {
	claudeCode AIProviderClient
	openAI     AIProviderClient
	ollama     AIProviderClient
	limiters   map[string]chan struct{}
}

func NewClient(openAIKey string, ollamaBaseURL, ollamaAuthToken string) *Client {
	return &Client{
		claudeCode: NewClaudeCodeClient(""),
		openAI:     NewOpenAIClient(openAIKey),
		ollama:     NewOllamaClient(ollamaBaseURL, ollamaAuthToken),
		limiters:   makeProviderLimiters(providerConcurrencyLimits),
	}
}

// IsAvailable reports whether the provider is locally configured for use.
func (c *Client) IsAvailable(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case ProviderClaudeCode:
		return c.claudeCode.IsAvailable()
	case ProviderOpenAI:
		return c.openAI.IsAvailable()
	case ProviderOllama:
		return c.ollama.IsAvailable()
	default:
		return false
	}
}

func (c *Client) Exec(provider string, model string, prompt string, attachmentPath string) (string, error) {
	return c.ExecWithLogLabel(provider, model, prompt, attachmentPath, "")
}

func (c *Client) ExecWithLogLabel(provider string, model string, prompt string, attachmentPath string, logLabel string) (string, error) {
	result, err := c.ExecResultWithLogLabel(provider, model, prompt, attachmentPath, logLabel)
	return result.Text, err
}

func (c *Client) ExecResultWithLogLabel(provider string, model string, prompt string, attachmentPath string, logLabel string) (Result, error) {
	return c.ExecPromptResultWithLogLabel(provider, model, Prompt{Dynamic: prompt}, attachmentPath, logLabel)
}

func (c *Client) ExecPromptResultWithLogLabel(provider string, model string, prompt Prompt, attachmentPath string, logLabel string) (Result, error) {
	normalizedProvider := strings.ToLower(strings.TrimSpace(provider))
	var clt AIProviderClient
	switch normalizedProvider {
	case ProviderClaudeCode:
		clt = c.claudeCode
	case ProviderOpenAI:
		clt = c.openAI
	case ProviderOllama:
		clt = c.ollama
	default:
		return Result{}, fmt.Errorf("llm exec: unsupported ai provider %q", provider)
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
	return clt.ExecPromptResultWithLogLabel(model, prompt, attachmentPath, logLabel)
}

type AIProviderClient interface {
	IsAvailable() bool
	ExecPromptResultWithLogLabel(model string, prompt Prompt, attachmentPath string, logLabel string) (Result, error)
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
