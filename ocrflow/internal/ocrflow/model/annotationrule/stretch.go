package annotationrule

type Stretch struct {
	Base            `json:",inline"`
	StretchCategory string      `json:"stretch_category"`
	Towards         string      `json:"towards"`
	ContactType     ContactType `json:"contact_type"`
	ContactSide     ContactSide `json:"contact_side"`
}

func (t *Stretch) GetType() Type {
	return TypeStretch
}

func (t *Stretch) SetDefaultValues() {
	t.StretchCategory = "MainZone-P"
	t.Towards = "bottom"
	t.ContactType = ContactTypeOuter
	t.ContactSide = ContactSideAll
}

func NewZeroStretch() *Stretch {
	return &Stretch{
		Base: Base{Type: TypeStretch, ApplicableStages: GetApplicableStages(TypeStretch)},
	}
}

func NewStretch(cat, towards string, ct ContactType, side ContactSide) *Stretch {
	s := NewZeroStretch()
	s.StretchCategory = cat
	s.Towards = towards
	s.ContactType = ct
	s.ContactSide = side
	return s
}
