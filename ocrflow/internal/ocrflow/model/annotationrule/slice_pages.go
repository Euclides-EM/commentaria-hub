package annotationrule

type SlicePages struct {
	Base  `json:",inline"`
	Pages string `json:"pages"`
}

func (t *SlicePages) GetType() Type {
	return TypeSlicePages
}

func NewSlicePages(pages string) *SlicePages {
	return &SlicePages{
		Base:  Base{Type: TypeSlicePages},
		Pages: pages,
	}
}
