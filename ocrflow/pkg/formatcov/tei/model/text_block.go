package model

type AB struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	Facs  string `xml:"facs,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`

	Segs []Seg `xml:"seg,omitempty"`
}
