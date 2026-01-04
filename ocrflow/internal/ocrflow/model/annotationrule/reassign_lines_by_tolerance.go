package annotationrule

type ReassignTextLinesByTolerance struct {
	Base         `json:",inline"`
	FromCategory string  `json:"from_category" example:"MainZone"`
	ToCategory   string  `json:"to_category" example:"MainZone-Head--Section"`
	PrecisionPx  float64 `json:"precision_px" example:"5.0"`
	// MinOverlap is the minimum overlap ratio (0.0 to 1.0) required to reassign a text line.
	// For example, a value of 0.8 means that at least 80% of the text line's width must overlap with the target category's bounding box to be reassigned.
	MinOverlap float64 `json:"min_overlap" example:"0.8"`
}

func (t *ReassignTextLinesByTolerance) GetType() Type {
	return TypeReassignTextLinesByTolerance
}

func NewReassignTextLinesByTolerance(fromCategory, toCategory string, precisionPx, minOverlap float64) *ReassignTextLinesByTolerance {
	return &ReassignTextLinesByTolerance{
		Base:         Base{Type: TypeReassignTextLinesByTolerance},
		FromCategory: fromCategory,
		ToCategory:   toCategory,
		PrecisionPx:  precisionPx,
		MinOverlap:   minOverlap,
	}
}
