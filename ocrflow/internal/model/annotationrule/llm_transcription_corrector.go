package annotationrule

type LLMTranscriptionCorrector struct {
	Base                        `json:",inline"`
	Provider                    string   `json:"provider" example:"ollama"`
	Model                       string   `json:"model" example:"gpt-oss:120b"`
	AdditionalAnnotations       []string `json:"additional_annotations"`
	IncludeEditionTranscription bool     `json:"include_edition_transcription"`
}

func (t *LLMTranscriptionCorrector) GetType() Type {
	return TypeLLMTranscriptionCorrector
}

func (t *LLMTranscriptionCorrector) SetDefaultValues() {
	t.Provider = "ollama"
	t.Model = "gpt-oss:120b"
	t.AdditionalAnnotations = []string{}
	t.IncludeEditionTranscription = false
}

func NewLLMTranscriptionCorrector(provider, model string, additionalAnnotations []string, includeEditionTranscription bool) *LLMTranscriptionCorrector {
	return &LLMTranscriptionCorrector{
		Base:                        Base{Type: TypeLLMTranscriptionCorrector, ApplicableStages: GetApplicableStages(TypeLLMTranscriptionCorrector)},
		Provider:                    provider,
		Model:                       model,
		AdditionalAnnotations:       additionalAnnotations,
		IncludeEditionTranscription: includeEditionTranscription,
	}
}
