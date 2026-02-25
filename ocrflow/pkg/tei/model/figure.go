package model

type Figure struct {
	XmlID   string   `xml:"xml:id,attr,omitempty"`
	Facs    string   `xml:"facs,attr,omitempty"`
	Graphic *Graphic `xml:"graphic,omitempty"`
	Desc    string   `xml:"figDesc,omitempty"`
}
