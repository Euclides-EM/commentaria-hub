package annotationrule

type Segment struct {
	Base  `json:",inline"`
	Model string `json:"model"`
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
