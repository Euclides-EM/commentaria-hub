package feature

type Result struct {
	DatasetID    string        `json:"dataset_id"`
	AnnotationID string        `json:"annotation_id"`
	Feature      string        `json:"feature"`
	Key          string        `json:"key"`
	Source       ResultSource  `json:"source"`
	Values       []ResultValue `json:"values"`

	Note string `json:"note,omitempty"`
}

type ResultSource struct {
	Resp     string `json:"resp"`
	Id       string `json:"id,omitempty"`
	Revision string `json:"revision,omitempty"`
	Name     string `json:"name,omitempty"`
}

type ResultValue struct {
	Root     string        `json:"root"`
	Children []ResultValue `json:"children,omitempty"`

	Source *ResultSource `json:"source,omitempty"`

	Note string `json:"note,omitempty"`
}
