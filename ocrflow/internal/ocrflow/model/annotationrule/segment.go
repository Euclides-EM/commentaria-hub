package annotationrule

type Segment struct {
	Base  `json:",inline"`
	Model string `json:"model" example:"1615FineTunedCapricciosaM_0312"`
}

func (t *Segment) GetType() Type {
	return TypeSegment
}

func NewSegment(model string) *Segment {
	return &Segment{
		Base:  Base{Type: TypeSegment},
		Model: model,
	}
}
