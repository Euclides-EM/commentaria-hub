package annotationrule

type AddMargin struct {
	Base     `json:",inline"`
	Margin   float64     `json:"margin"`
	Side     ContactSide `json:"sides"`
	Category string      `json:"category"`
}

func (t *AddMargin) GetType() Type {
	return TypeAddMargin
}

func NewAddMargin(margin float64, side ContactSide) *AddMargin {
	return &AddMargin{
		Base:   Base{Type: TypeAddMargin},
		Margin: margin,
		Side:   side,
	}
}
