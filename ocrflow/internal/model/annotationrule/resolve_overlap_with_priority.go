package annotationrule

// ResolveOverlapWithPriority removes zones of SuppressedCategory that overlap
// zones of DominantCategory by at least MinOverlap percent (of the suppressed zone area).
type ResolveOverlapWithPriority struct {
	Base               `json:",inline"`
	DominantCategory   string  `json:"dominant_category" example:"MainZone"`
	SuppressedCategory string  `json:"suppressed_category" example:"DigitizationArtefactZone"`
	MinOverlap         float64 `json:"min_overlap" example:"0.8"` // percentage 0–1
}

func (r *ResolveOverlapWithPriority) GetType() Type {
	return TypeResolveOverlapWithPriority
}

func (r *ResolveOverlapWithPriority) SetDefaultValues() {
	r.DominantCategory = "RunningTitleZone"
	r.SuppressedCategory = "MainZone-Head--Section"
	r.MinOverlap = 0.8
}

func NewResolveOverlapWithPriority(dominantCategory, suppressedCategory string, minOverlapPct float64) *ResolveOverlapWithPriority {
	return &ResolveOverlapWithPriority{
		Base:               Base{Type: TypeResolveOverlapWithPriority, ApplicableStages: GetApplicableStages(TypeResolveOverlapWithPriority)},
		DominantCategory:   dominantCategory,
		SuppressedCategory: suppressedCategory,
		MinOverlap:         minOverlapPct,
	}
}
