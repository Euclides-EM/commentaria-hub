package service

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
)

type stubLLMAvailability map[string]bool

func (a stubLLMAvailability) IsAvailable(provider string) bool {
	return a[provider]
}

func TestSuggestedLLMTranscriptionCorrectorUsesAvailableProviderFallback(t *testing.T) {
	tests := []struct {
		name      string
		available stubLLMAvailability
		provider  string
		model     string
	}{
		{
			name:      "claude code preferred",
			available: stubLLMAvailability{llm.ProviderClaudeCode: true, llm.ProviderOpenAI: true, llm.ProviderOllama: true},
			provider:  llm.ProviderClaudeCode,
			model:     "fable",
		},
		{
			name:      "openai fallback",
			available: stubLLMAvailability{llm.ProviderOpenAI: true, llm.ProviderOllama: true},
			provider:  llm.ProviderOpenAI,
			model:     "gpt-5.6-terra",
		},
		{
			name:      "ollama fallback",
			available: stubLLMAvailability{llm.ProviderOllama: true},
			provider:  llm.ProviderOllama,
			model:     "qwen3-vl:32b",
		},
		{
			name:      "ollama remains final suggestion when none configured",
			available: stubLLMAvailability{},
			provider:  llm.ProviderOllama,
			model:     "qwen3-vl:32b",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataset := &Dataset{llmAvailability: test.available}
			provider, model := dataset.suggestedLLMTranscriptionCorrector()
			if provider != test.provider || model != test.model {
				t.Fatalf("provider/model = %s/%s, want %s/%s", provider, model, test.provider, test.model)
			}
		})
	}
}
