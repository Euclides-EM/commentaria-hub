package model

type Seg struct {
	XmlID   string   `xml:"xml:id,attr,omitempty"`
	Corresp string   `xml:"corresp,attr,omitempty"`
	Facs    string   `xml:"facs,attr,omitempty"`
	Content []Inline `xml:",any,omitempty"`
}
