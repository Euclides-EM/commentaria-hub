package model

type Annotation struct {
	Meta     `json:",inline"`
	Pages    string    `json:"pages"`
	LocalDir string    `json:"local_dir" readonly:"true"`
	Dataset  Reference `json:"dataset" readonly:"true"`
	Model    Reference `json:"model"`
}

func (a *Annotation) DatasetID() string {
	return a.Dataset.ID
}

func (a *Annotation) ModelID() string {
	return a.Model.ID
}

func (a *Annotation) DeepCopy() *Annotation {
	if a == nil {
		return nil
	}
	return &Annotation{
		Meta:    a.Meta.DeepCopy(),
		Pages:   a.Pages,
		Dataset: a.Dataset.DeepCopy(),
		Model:   a.Model.DeepCopy(),
	}
}
