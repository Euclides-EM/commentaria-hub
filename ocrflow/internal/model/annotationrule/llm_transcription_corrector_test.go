package annotationrule

import (
	"encoding/json"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalLLMTranscriptionCorrector(t *testing.T) {
	require.Contains(t, AllAnnotationRuleTypes, TypeLLMTranscriptionCorrector)
	require.Contains(t, PreferAsyncTypes, TypeLLMTranscriptionCorrector)

	rule, err := UnmarshalRuleJSON([]byte(`{
		"type":"llm_transcription_corrector",
		"provider":"openai",
		"model":"gpt-test",
		"rounds":3,
		"pages":"2-4,7",
		"skip_existing":true,
		"additional_annotations":["ann_one","ann_two"],
		"include_edition_transcription":true
	}`))
	require.NoError(t, err)

	corrector, ok := rule.(*LLMTranscriptionCorrector)
	require.True(t, ok)
	require.Equal(t, "openai", corrector.Provider)
	require.Equal(t, "gpt-test", corrector.Model)
	require.Equal(t, 3, corrector.Rounds)
	require.Equal(t, "2-4,7", corrector.Pages)
	require.True(t, corrector.SkipExisting)
	require.Equal(t, []string{"ann_one", "ann_two"}, corrector.AdditionalAnnotations)
	require.True(t, corrector.IncludeEditionTranscription)
	require.Equal(t, []PipelineStage{PipelineStageOCR}, corrector.ApplicableStages)

	encoded, err := json.Marshal(corrector)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"type":"llm_transcription_corrector"`)
	require.Contains(t, string(encoded), `"rounds":3`)
	require.Contains(t, string(encoded), `"pages":"2-4,7"`)
	require.Contains(t, string(encoded), `"skip_existing":true`)
	require.NotContains(t, string(encoded), `"usage"`)
}

func TestLLMTranscriptionCorrectorUsageJSONRoundTrip(t *testing.T) {
	cost := 0.125
	rule := NewLLMTranscriptionCorrector("claude-code", "fable", nil, false)
	rule.Usage = &llm.Usage{
		InputTokens: 100, CachedInputTokens: 20, CacheCreationInputTokens: 10,
		CacheMetricsAvailable: true, OutputTokens: 30, TotalTokens: 160, CostUSD: &cost,
	}

	encoded, err := json.Marshal(rule)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"usage":{"input_tokens":100,"cached_input_tokens":20,"cache_creation_input_tokens":10,"cache_metrics_available":true,"output_tokens":30,"reasoning_tokens":0,"total_tokens":160,"cost_usd":0.125}`)

	decoded, err := UnmarshalRuleJSON(encoded)
	require.NoError(t, err)
	corrector := decoded.(*LLMTranscriptionCorrector)
	require.Equal(t, rule.Usage, corrector.Usage)
}

func TestLLMTranscriptionCorrectorDefaultsRounds(t *testing.T) {
	rule := NewLLMTranscriptionCorrector("ollama", "vision", nil, false)
	require.Equal(t, 1, rule.Rounds)

	rule.SetDefaultValues()
	require.Equal(t, 1, rule.Rounds)
	require.False(t, rule.SkipExisting)
}
