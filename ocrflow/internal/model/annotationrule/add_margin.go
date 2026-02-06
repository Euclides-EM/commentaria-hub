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

func (t *AddMargin) SetDefaultValues() {
	t.Margin = 2.0
	t.Side = ContactSideAll
	t.Category = "MainZone-Head--Section"
}

func NewZeroAddMargin() *AddMargin {
	return &AddMargin{
		Base: Base{Type: TypeAddMargin, ApplicableStages: GetApplicableStages(TypeAddMargin)},
	}
}

func NewAddMargin(margin float64, side ContactSide) *AddMargin {
	am := NewZeroAddMargin()
	am.Margin = margin
	am.Side = side
	return am
}
