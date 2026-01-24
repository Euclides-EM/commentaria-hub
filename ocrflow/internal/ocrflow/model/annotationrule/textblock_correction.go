package annotationrule

type TextBlockCorrections struct {
	Base        `json:",inline"`
	Corrections []*TextBlockCorrection `json:"corrections"`
}

func (t *TextBlockCorrections) GetType() Type {
	return TypeTextBlocksCorrections
}

func (t *TextBlockCorrections) SetDefaultValues() {
	t.Corrections = []*TextBlockCorrection{
		NewTextBlockCorrection("example_text_block_id_1", []string{"Corrected text for block 1."}),
		NewTextBlockCorrection("example_text_block_id_2", []string{"Corrected text for block 2."}),
	}
}

func NewTextBlockCorrections(corrections []*TextBlockCorrection) *TextBlockCorrections {
	return &TextBlockCorrections{
		Base:        Base{Type: TypeTextBlocksCorrections, ApplicableStages: GetApplicableStages(TypeTextBlocksCorrections)},
		Corrections: corrections,
	}
}

type TextBlockCorrection struct {
	Page        int      `json:"page"`
	TextBlockID string   `json:"text_block_id"`
	Correction  []string `json:"correction"`
}

func NewTextBlockCorrection(textBlockID string, correction []string) *TextBlockCorrection {
	return &TextBlockCorrection{
		TextBlockID: textBlockID,
		Correction:  correction,
	}
}
