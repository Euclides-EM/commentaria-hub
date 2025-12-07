package annotationrule

type RemoveCategories struct {
	Base       `json:",inline"`
	Categories []string `json:"categories" example:""`
}

func (t *RemoveCategories) GetType() Type {
	return TypeRemoveCategories
}

func NewRemoveCategories(categories []string) *RemoveCategories {
	return &RemoveCategories{
		Base:       Base{Type: TypeRemoveCategories},
		Categories: categories,
	}
}
