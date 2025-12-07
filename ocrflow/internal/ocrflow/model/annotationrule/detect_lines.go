package annotationrule

type LinesDetect struct {
	Base `json:",inline"`
	// IncludeCategories specifies which categories to run line detection on. For example, "MainZone".
	// Example: ["MainZone"]
	IncludeCategories []string `json:"include_categories,omitempty"`
	// IgnoreCategories specifies which categories to ignore when running line detection. For example, "GraphicZone", "DigitizationArtefactZone", ...
	// Example: ["CatchWord", "DigitizationArtefactZone", "DropCapitalZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "NumberingZone", "QuireMarksZone", "RunningTitleZone"]
	IgnoreCategories []string `json:"ignore_categories,omitempty"`
}

func (t *LinesDetect) GetType() Type {
	return TypeLinesDetect
}

func NewLinesDetect(includeCategories, ignoreCategories []string) *LinesDetect {
	return &LinesDetect{
		Base:              Base{Type: TypeLinesDetect},
		IncludeCategories: includeCategories,
		IgnoreCategories:  ignoreCategories,
	}
}
