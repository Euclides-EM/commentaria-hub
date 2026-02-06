package featureplat

type FeatureResult struct {
	Feature string               `json:"feature"`
	Key     string               `json:"key"`
	Source  FeatureResultSource  `json:"source"`
	Values  []FeatureResultValue `json:"values"`

	Note string `json:"note,omitempty"`
}

type FeatureResultSource struct {
	Resp     string `json:"resp"`
	Id       string `json:"id,omitempty"`
	Revision string `json:"revision,omitempty"`
	Name     string `json:"name,omitempty"`
}

type FeatureResultValue struct {
	Root      string               `json:"root"`
	Childrens []FeatureResultValue `json:"childrens,omitempty"`

	Source *FeatureResultSource `json:"source,omitempty"`

	Note string `json:"note,omitempty"`
}
