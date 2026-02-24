package model

import "encoding/xml"

type Inline struct {
	XMLName xml.Name

	Text string `xml:",chardata"`

	Corresp string `xml:"corresp,attr,omitempty"`
	Facs    string `xml:"facs,attr,omitempty"`
	Ref     string `xml:"ref,attr,omitempty"`
	Ana     string `xml:"ana,attr,omitempty"`
}
