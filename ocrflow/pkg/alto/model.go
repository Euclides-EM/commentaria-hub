package alto

import "encoding/xml"

type Alto struct {
	XMLName        xml.Name    `xml:"alto"`
	Xmlns          string      `xml:"xmlns,attr,omitempty"`
	XmlnsXsi       string      `xml:"xmlns:xsi,attr,omitempty"`
	SchemaLocation string      `xml:"xsi:schemaLocation,attr,omitempty"`
	Description    Description `xml:"Description"`
	Tags           Tags        `xml:"Tags"`
	Layout         Layout      `xml:"Layout"`
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
	Page []Page `xml:"Page"`
}

type Page struct {
	Width      int        `xml:"WIDTH,attr"`
	Height     int        `xml:"HEIGHT,attr"`
	ID         string     `xml:"ID,attr"`
	PrintSpace PrintSpace `xml:"PrintSpace"`
}

type PrintSpace struct {
	HPOS       float64     `xml:"HPOS,attr"`
	VPOS       float64     `xml:"VPOS,attr"`
	Width      float64     `xml:"WIDTH,attr"`
	Height     float64     `xml:"HEIGHT,attr"`
	TextBlocks []TextBlock `xml:"TextBlock"`
}

type TextBlock struct {
	ID      string     `xml:"ID,attr"`
	TagRefs string     `xml:"TAGREFS,attr"`
	HPOS    float64    `xml:"HPOS,attr"`
	VPOS    float64    `xml:"VPOS,attr"`
	Width   float64    `xml:"WIDTH,attr"`
	Height  float64    `xml:"HEIGHT,attr"`
	Shape   Shape      `xml:"Shape"`
	Lines   []TextLine `xml:"TextLine"`
}

type TextLine struct {
	ID     string  `xml:"ID,attr"`
	HPOS   float64 `xml:"HPOS,attr"`
	VPOS   float64 `xml:"VPOS,attr"`
	Width  float64 `xml:"WIDTH,attr"`
	Height float64 `xml:"HEIGHT,attr"`
	// child elements
	Strings []AltoString `xml:"http://www.loc.gov/standards/alto/ns-v4# String"`
}

type AltoString struct {
	// Change this tag to match the XML exactly
	Content string  `xml:"CONTENT,attr"`
	HPOS    float64 `xml:"HPOS,attr,omitempty"`
	VPOS    float64 `xml:"VPOS,attr,omitempty"`
	Width   float64 `xml:"WIDTH,attr,omitempty"`
	Height  float64 `xml:"HEIGHT,attr,omitempty"`
	WC      float64 `xml:"WC,attr,omitempty"`
}

type Line struct {
	BlockID  string
	TagRefs  string
	HPOS     float64
	VPOS     float64
	Strings  []AltoString
	Height   float64
	LineID   string
	BlockVP  float64
	BlockHP  float64
	BlockHgt float64
}

func (s *Line) Text() string {
	text := ""
	for _, str := range s.Strings {
		text += str.Content
	}
	return text
}

type Description struct {
	MeasurementUnit        string                 `xml:"MeasurementUnit"`
	SourceImageInformation SourceImageInformation `xml:"sourceImageInformation"`
}

type SourceImageInformation struct {
	FileName string `xml:"fileName"`
}

type Shape struct {
	Polygon Polygon `xml:"Polygon"`
}

type Polygon struct {
	Points string `xml:"POINTS,attr"`
}
