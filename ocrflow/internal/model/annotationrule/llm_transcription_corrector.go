package annotationrule

import "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/transcriptioncorrector"

type LLMTranscriptionCorrector struct {
	Base                        `json:",inline"`
	Provider                    string   `json:"provider" example:"ollama"`
	Model                       string   `json:"model" example:"gpt-oss:120b"`
	Rounds                      int      `json:"rounds" example:"1" minimum:"1"`
	AdditionalAnnotations       []string `json:"additional_annotations"`
	IncludeEditionTranscription bool     `json:"include_edition_transcription"`
}

func (t *LLMTranscriptionCorrector) GetType() Type {
	return TypeLLMTranscriptionCorrector
}

func (t *LLMTranscriptionCorrector) SetDefaultValues() {
	t.Provider = "ollama"
	t.Model = "gpt-oss:120b"
	t.Rounds = transcriptioncorrector.DefaultRounds
	t.AdditionalAnnotations = []string{}
	t.IncludeEditionTranscription = false
}

func NewLLMTranscriptionCorrector(provider, model string, additionalAnnotations []string, includeEditionTranscription bool) *LLMTranscriptionCorrector {
	return &LLMTranscriptionCorrector{
		Base:                        Base{Type: TypeLLMTranscriptionCorrector, ApplicableStages: GetApplicableStages(TypeLLMTranscriptionCorrector)},
		Provider:                    provider,
		Model:                       model,
		Rounds:                      transcriptioncorrector.DefaultRounds,
		AdditionalAnnotations:       additionalAnnotations,
		IncludeEditionTranscription: includeEditionTranscription,
	}
}
