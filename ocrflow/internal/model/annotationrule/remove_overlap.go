package annotationrule

type RemoveOverlap struct {
	Base       `json:",inline"`
	Categories []string `json:"categories"`
	Precision  float64  `json:"precision" example:"1000"`
}

func (r *RemoveOverlap) GetType() Type {
	return TypeRemoveOverlap
}

func (r *RemoveOverlap) SetDefaultValues() {
	r.Categories = []string{"DigitizationArtefactZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "MainZone", "MainZone-Head--Book", "MainZone-Head--Section", "NumberingZone", "QuireMarksZone", "RunningTitleZone"}
	r.Precision = 1000.0
}

func NewRemoveOverlap(categories []string, precision float64) *RemoveOverlap {
	return &RemoveOverlap{
		Base:       Base{Type: TypeRemoveOverlap, ApplicableStages: GetApplicableStages(TypeRemoveOverlap)},
		Categories: categories,
		Precision:  precision,
	}
}
