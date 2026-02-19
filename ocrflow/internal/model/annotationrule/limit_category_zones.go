package annotationrule

// KeepPosition indicates which zones to keep when limiting by count (closest to that page side).
type KeepPosition string

const (
	KeepPositionTop    KeepPosition = "top"
	KeepPositionBottom KeepPosition = "bottom"
	KeepPositionLeft   KeepPosition = "left"
	KeepPositionRight  KeepPosition = "right"
)

// LimitCategoryZones ensures at most MaxCount zones exist for Category per page.
// If more are present, keeps only the ones closest to KeepPosition (top/bottom/left/right)
// and removes the rest.
type LimitCategoryZones struct {
	Base         `json:",inline"`
	Category     string       `json:"category" example:"RunningTitleZone"`
	MaxCount     int          `json:"max_count" example:"1"`
	KeepPosition KeepPosition `json:"keep_position" example:"top"`
}

func (r *LimitCategoryZones) GetType() Type {
	return TypeLimitCategoryZones
}

func (r *LimitCategoryZones) SetDefaultValues() {
	r.Category = "RunningTitleZone"
	r.MaxCount = 1
	r.KeepPosition = KeepPositionTop
}

func NewLimitCategoryZones(category string, maxCount int, keepPosition KeepPosition) *LimitCategoryZones {
	return &LimitCategoryZones{
		Base:         Base{Type: TypeLimitCategoryZones, ApplicableStages: GetApplicableStages(TypeLimitCategoryZones)},
		Category:     category,
		MaxCount:     maxCount,
		KeepPosition: keepPosition,
	}
}
