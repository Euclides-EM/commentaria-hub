package models

type Annotation struct {
	EditionKey        string `json:"edition_key"`
	FacsimileId       string `json:"facsimile_id"`
	Model             string `json:"model"`
	Pages             string `json:"pages"`
	ID                string `json:"id"`
	AnnotatedLocalDir string `json:"annotated_local_dir"`
}

func (a *Annotation) DeepCopy() *Annotation {
	if a == nil {
		return nil
	}
	return &Annotation{
		EditionKey:        a.EditionKey,
		FacsimileId:       a.FacsimileId,
		Model:             a.Model,
		Pages:             a.Pages,
		ID:                a.ID,
		AnnotatedLocalDir: a.AnnotatedLocalDir,
	}
}
