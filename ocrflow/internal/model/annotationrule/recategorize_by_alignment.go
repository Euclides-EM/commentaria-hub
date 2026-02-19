package annotationrule

type Alignment string

const (
	AlignmentHorizontal Alignment = "horizontal"
	AlignmentVertical   Alignment = "vertical"
)

// RecategorizeByAlignmentRelativeTo defines the reference category and alignment axis.
type RecategorizeByAlignmentRelativeTo struct {
	Category  string    `json:"category" example:"Pagination"`
	Alignment Alignment `json:"alignment" example:"horizontal"` // "horizontal" or "vertical"
}

// RecategorizeByAlignment finds zones with original_category that are horizontally or
// vertically aligned (within tolerance_px) with zones of relative_to.category, and
// changes their category to target_category. No overlap required.
type RecategorizeByAlignment struct {
	Base             `json:",inline"`
	OriginalCategory string                            `json:"original_category" example:"MainZone-Head--Section"`
	TargetCategory   string                            `json:"target_category" example:"RunningTitleZone"`
	RelativeTo       RecategorizeByAlignmentRelativeTo `json:"relative_to"`
	TolerancePx      float64                           `json:"tolerance_px" example:"8.0"`
}

func (r *RecategorizeByAlignment) GetType() Type {
	return TypeRecategorizeByAlignment
}

func (r *RecategorizeByAlignment) SetDefaultValues() {
	r.OriginalCategory = "MainZone-Head--Section"
	r.TargetCategory = "RunningTitleZone"
	r.RelativeTo = RecategorizeByAlignmentRelativeTo{
		Category:  "Pagination",
		Alignment: AlignmentHorizontal,
	}
	r.TolerancePx = 8.0
}

func NewRecategorizeByAlignment(originalCategory, targetCategory, relativeCategory string, alignment Alignment, tolerancePx float64) *RecategorizeByAlignment {
	return &RecategorizeByAlignment{
		Base:             Base{Type: TypeRecategorizeByAlignment, ApplicableStages: GetApplicableStages(TypeRecategorizeByAlignment)},
		OriginalCategory: originalCategory,
		TargetCategory:   targetCategory,
		RelativeTo:       RecategorizeByAlignmentRelativeTo{Category: relativeCategory, Alignment: alignment},
		TolerancePx:      tolerancePx,
	}
}
