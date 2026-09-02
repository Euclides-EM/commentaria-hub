package annotationrule

import (
	"encoding/json"
	"testing"

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
		"additional_annotations":["ann_one","ann_two"],
		"include_edition_transcription":true
	}`))
	require.NoError(t, err)

	corrector, ok := rule.(*LLMTranscriptionCorrector)
	require.True(t, ok)
	require.Equal(t, "openai", corrector.Provider)
	require.Equal(t, "gpt-test", corrector.Model)
	require.Equal(t, 3, corrector.Rounds)
	require.Equal(t, []string{"ann_one", "ann_two"}, corrector.AdditionalAnnotations)
	require.True(t, corrector.IncludeEditionTranscription)
	require.Equal(t, []PipelineStage{PipelineStageOCR}, corrector.ApplicableStages)

	encoded, err := json.Marshal(corrector)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"type":"llm_transcription_corrector"`)
	require.Contains(t, string(encoded), `"rounds":3`)
}

func TestLLMTranscriptionCorrectorDefaultsRounds(t *testing.T) {
	rule := NewLLMTranscriptionCorrector("ollama", "vision", nil, false)
	require.Equal(t, 1, rule.Rounds)

	rule.SetDefaultValues()
	require.Equal(t, 1, rule.Rounds)
}
