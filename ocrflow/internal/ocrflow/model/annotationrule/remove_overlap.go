package annotationrule

type RemoveOverlap struct {
	Base       `json:",inline"`
	Categories []string `json:"categories"`
	Precision  float64  `json:"precision" example:"1000"`
}

func (r *RemoveOverlap) GetType() Type {
	return TypeRemoveOverlap
}

func NewRemoveOverlap(categories []string, precision float64) *RemoveOverlap {
	return &RemoveOverlap{
		Base:       Base{Type: TypeRemoveOverlap},
		Categories: categories,
		Precision:  precision,
	}
}
