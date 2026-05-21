package feature

import "strings"

type AIProvider string

func (p AIProvider) ToLLMAIProvider() string {
	return strings.ToLower(string(p))
}

const (
	AIProviderOllama AIProvider = "ollama"
	AIProviderOpenAI AIProvider = "openai"
)
