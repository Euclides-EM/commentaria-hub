package models

type Edition struct {
	Key        string       `json:"key"`
	Facsimiles []*Facsimile `json:"facsimiles,omitempty"`
}
