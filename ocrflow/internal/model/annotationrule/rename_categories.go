package annotationrule

type RenameCategories struct {
	Base    `json:",inline"`
	Renames map[string]string `json:"renames" example:""`
}

func (r *RenameCategories) GetType() Type {
	return TypeRenameCategories
}

func (r *RenameCategories) SetDefaultValues() {
	r.Renames = map[string]string{
		"DropCapitalZone-Plane": "DropCapitalZone-Plain",
	}
}

func NewRenameCategories(renames map[string]string) *RenameCategories {
	cloned := make(map[string]string, len(renames))
	for from, to := range renames {
		cloned[from] = to
	}
	return &RenameCategories{
		Base:    Base{Type: TypeRenameCategories, ApplicableStages: GetApplicableStages(TypeRenameCategories)},
		Renames: cloned,
	}
}
