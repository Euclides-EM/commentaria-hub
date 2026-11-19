package annotationrule

import (
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/transcriptioncorrector"
)

type LLMTranscriptionCorrector struct {
	Base     `json:",inline"`
	Provider string `json:"provider" example:"ollama"`
	Model    string `json:"model" example:"gpt-oss:120b"`
	Rounds   int    `json:"rounds" example:"1"`
	// ExecutionMode selects per-page requests or one local-agent directory request. Directory mode ignores Rounds.
	ExecutionMode               transcriptioncorrector.ExecutionMode `json:"execution_mode" swaggertype:"string" enums:"page_by_page,directory" example:"page_by_page"`
	SkipExisting                bool                                 `json:"skip_existing"`
	AdditionalAnnotations       []string                             `json:"additional_annotations"`
	IncludeEditionTranscription bool                                 `json:"include_edition_transcription"`
	Usage                       *llm.Usage                           `json:"usage,omitempty" readonly:"true"`
}

func (t *LLMTranscriptionCorrector) GetType() Type {
	return TypeLLMTranscriptionCorrector
}

func (t *LLMTranscriptionCorrector) SetDefaultValues() {
	t.Provider = "ollama"
	t.Model = "gpt-oss:120b"
	t.Rounds = transcriptioncorrector.DefaultRounds
	t.ExecutionMode = transcriptioncorrector.DefaultExecutionMode
	t.SkipExisting = false
	t.AdditionalAnnotations = []string{}
	t.IncludeEditionTranscription = false
}

func NewLLMTranscriptionCorrector(provider, model string, additionalAnnotations []string, includeEditionTranscription bool) *LLMTranscriptionCorrector {
	return &LLMTranscriptionCorrector{
		Base:                        Base{Type: TypeLLMTranscriptionCorrector, ApplicableStages: GetApplicableStages(TypeLLMTranscriptionCorrector)},
		Provider:                    provider,
		Model:                       model,
		Rounds:                      transcriptioncorrector.DefaultRounds,
		ExecutionMode:               transcriptioncorrector.DefaultExecutionMode,
		AdditionalAnnotations:       additionalAnnotations,
		IncludeEditionTranscription: includeEditionTranscription,
	}
}
