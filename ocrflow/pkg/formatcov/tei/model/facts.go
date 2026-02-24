package model

type StandOff struct {
	Relations []Relation `xml:"listRelation>relation,omitempty"`
}

type Relation struct {
	XmlID   string `xml:"xml:id,attr,omitempty"`
	Name    string `xml:"name,attr"`
	Active  string `xml:"active,attr,omitempty"`
	Passive string `xml:"passive,attr,omitempty"`
	Ana     string `xml:"ana,attr,omitempty"`
	Cert    string `xml:"cert,attr,omitempty"`
}
