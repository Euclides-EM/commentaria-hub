package alto

import "encoding/xml"

type Alto struct {
	XMLName xml.Name `xml:"alto"`
	Tags    Tags     `xml:"Tags"`
	Layout  Layout   `xml:"Layout"`
}

type Tags struct {
	OtherTags []OtherTag `xml:"OtherTag"`
}

type OtherTag struct {
	ID          string `xml:"ID,attr"`
	Label       string `xml:"LABEL,attr"`
	Description string `xml:"DESCRIPTION,attr"`
}

type Layout struct {
	Pages []Page `xml:"Page"`
}

type Page struct {
	Width      int        `xml:"WIDTH,attr"`
	Height     int        `xml:"HEIGHT,attr"`
	ID         string     `xml:"ID,attr"`
	PrintSpace PrintSpace `xml:"PrintSpace"`
}

type PrintSpace struct {
	TextBlocks []TextBlock `xml:"TextBlock"`
}

type TextBlock struct {
	ID      string `xml:"ID,attr"`
	TagRefs string `xml:"TAGREFS,attr"`

	HPOS   int `xml:"HPOS,attr"`
	VPOS   int `xml:"VPOS,attr"`
	Width  int `xml:"WIDTH,attr"`
	Height int `xml:"HEIGHT,attr"`
}
