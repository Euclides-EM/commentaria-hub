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
		"additional_annotations":["ann_one","ann_two"],
		"include_edition_transcription":true
	}`))
	require.NoError(t, err)

	corrector, ok := rule.(*LLMTranscriptionCorrector)
	require.True(t, ok)
	require.Equal(t, "openai", corrector.Provider)
	require.Equal(t, "gpt-test", corrector.Model)
	require.Equal(t, []string{"ann_one", "ann_two"}, corrector.AdditionalAnnotations)
	require.True(t, corrector.IncludeEditionTranscription)
	require.Equal(t, []PipelineStage{PipelineStageOCR}, corrector.ApplicableStages)

	encoded, err := json.Marshal(corrector)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"type":"llm_transcription_corrector"`)
}
