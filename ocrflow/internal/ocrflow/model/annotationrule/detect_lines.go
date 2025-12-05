package annotationrule

type LinesDetect struct {
	Base `json:",inline"`
	// IncludeCategories specifies which categories to run line detection on. For example, "MainZone".
	IncludeCategories []string `json:"include_categories,omitempty"`
	// IgnoreCategories specifies which categories to ignore when running line detection. For example, "GraphicZone", "DigitizationArtefactZone", ...
	IgnoreCategories []string `json:"ignore_categories,omitempty"`
}

func (t *LinesDetect) GetType() Type {
	return TypeLinesDetect
}

func NewLinesDetect(detect bool) *LinesDetect {
	return &LinesDetect{
		Base: Base{Type: TypeLinesDetect},
	}
}
