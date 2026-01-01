package annotationrule

type SlicePages struct {
	Base        `json:",inline"`
	Pages       string `json:"pages"`
	RandomPages int    `json:"random_pages,omitempty"`
}

func (t *SlicePages) GetType() Type {
	return TypeSlicePages
}

func NewSlicePagesFixed(pages string) *SlicePages {
	return &SlicePages{
		Base:  Base{Type: TypeSlicePages},
		Pages: pages,
	}
}
