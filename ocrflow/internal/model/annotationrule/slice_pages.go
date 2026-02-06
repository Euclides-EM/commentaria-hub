package annotationrule

type SlicePages struct {
	Base        `json:",inline"`
	Pages       string `json:"pages"`
	RandomPages int    `json:"random_pages,omitempty"`
}

func (t *SlicePages) GetType() Type {
	return TypeSlicePages
}

func (t *SlicePages) SetDefaultValues() {
	t.Pages = "1-5"
	t.RandomPages = 0
}

func NewSlicePagesFixed(pages string) *SlicePages {
	return &SlicePages{
		Base:  Base{Type: TypeSlicePages, ApplicableStages: GetApplicableStages(TypeSlicePages)},
		Pages: pages,
	}
}
