package model

type Text struct {
	Body Body `xml:"body"`
}

type Body struct {
	PBs     []PB     `xml:"pb,omitempty"`
	Figures []Figure `xml:"figure,omitempty"`
	Divs    []Div    `xml:"div,omitempty"`
}
