package annotationrule

type RemoveCategories struct {
	Base       `json:",inline"`
	Categories []string `json:"categories" example:""`
}

func (t *RemoveCategories) GetType() Type {
	return TypeRemoveCategories
}

func (t *RemoveCategories) SetDefaultValues() {
	t.Categories = []string{"MainZone-P--Italics", "MainZone-P--Enunciation", "MainZone-P"}
}

func NewRemoveCategories(categories []string) *RemoveCategories {
	return &RemoveCategories{
		Base:       Base{Type: TypeRemoveCategories, ApplicableStages: GetApplicableStages(TypeRemoveCategories)},
		Categories: categories,
	}
}
