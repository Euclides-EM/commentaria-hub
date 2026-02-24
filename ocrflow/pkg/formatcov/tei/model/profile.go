package model

type ProfileDesc struct {
	TextClass *TextClass `xml:"textClass,omitempty"`
}

type TextClass struct {
	Keywords *Keywords `xml:"keywords,omitempty"`
}

type Keywords struct {
	Scheme string `xml:"scheme,attr,omitempty"`
	Terms  []Term `xml:"term,omitempty"`
}

type Term struct {
	// Generic TEI term; used for profile key-values.
	Type    string `xml:"type,attr,omitempty"`
	Corresp string `xml:"corresp,attr,omitempty"`
	Text    string `xml:",chardata"`
}
