package annotationrule

type TextBlockCorrections struct {
	Base        `json:",inline"`
	Corrections []*TextBlockCorrection `json:"corrections"`
}

func (t *TextBlockCorrections) GetType() Type {
	return TypeTextBlocksCorrections
}

type TextBlockCorrection struct {
	Base        `json:",inline"`
	Page        int      `json:"page"`
	TextBlockID string   `json:"text_block_id"`
	Correction  []string `json:"correction"`
}

func NewTextBlockCorrection(textBlockID string, correction []string) *TextBlockCorrection {
	return &TextBlockCorrection{
		Base:        Base{Type: TypeTextBlocksCorrections},
		TextBlockID: textBlockID,
		Correction:  correction,
	}
}
