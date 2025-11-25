package model

type Reference struct {
	ID string `json:"id"`
}

func (r Reference) DeepCopy() Reference {
	return Reference{
		ID: r.ID,
	}
}
