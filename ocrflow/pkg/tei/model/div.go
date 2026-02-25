package model

type Div struct {
	Type    string `xml:"type,attr,omitempty"`
	XmlLang string `xml:"xml:lang,attr,omitempty"`
	N       string `xml:"n,attr,omitempty"`
	Abs     []AB   `xml:"ab,omitempty"`
}
