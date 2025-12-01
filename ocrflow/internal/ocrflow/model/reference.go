package model

type Reference struct {
	ID string `json:"id"`
}

func (r Reference) DeepCopy() Reference {
	return Reference{
		ID: r.ID,
	}
}

type AnnotationReference struct {
	ID        string `json:"id"`
	DatasetId string `json:"dataset_id"`
}

func (ar AnnotationReference) DeepCopy() AnnotationReference {
	return AnnotationReference{
		ID:        ar.ID,
		DatasetId: ar.DatasetId,
	}
}
