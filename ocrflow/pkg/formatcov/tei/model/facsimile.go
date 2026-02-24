package model

type Facsimile struct {
	Surfaces []Surface `xml:"surface"`
}

type Surface struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	Facs  string `xml:"facs,attr,omitempty"` // page image
	N     string `xml:"n,attr,omitempty"`

	Zones []Zone `xml:"zone"`
}

type Zone struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`

	ULX float64 `xml:"ulx,attr,omitempty"`
	ULY float64 `xml:"uly,attr,omitempty"`
	LRX float64 `xml:"lrx,attr,omitempty"`
	LRY float64 `xml:"lry,attr,omitempty"`

	Corresp string `xml:"corresp,attr,omitempty"`

	Graphic *Graphic `xml:"graphic,omitempty"` // diagrams
}

type Graphic struct {
	URL string `xml:"url,attr"`
}
