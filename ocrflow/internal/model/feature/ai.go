package feature

import "strings"

type AIProvider string

func (p AIProvider) ToLLMAIProvider() string {
	return strings.ToLower(string(p))
}

const (
	AIProviderClaudeCode AIProvider = "claude-code"
	AIProviderOllama     AIProvider = "ollama"
	AIProviderOpenAI     AIProvider = "openai"
)
