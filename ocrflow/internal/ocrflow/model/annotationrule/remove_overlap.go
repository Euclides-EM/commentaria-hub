package annotationrule

type RemoveOverlap struct {
	Base       `json:",inline"`
	Categories []string `json:"categories"`
	Precision  int      `json:"precision" example:"1000"`
}

func (r *RemoveOverlap) GetType() Type {
	return TypeRemoveOverlap
}

func NewRemoveOverlap(categories []string, precision int) *RemoveOverlap {
	return &RemoveOverlap{
		Base:       Base{Type: TypeRemoveOverlap},
		Categories: categories,
		Precision:  precision,
	}
}
