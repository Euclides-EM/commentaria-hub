package model

import "encoding/xml"

type L struct {
	XmlID   string `xml:"xml:id,attr,omitempty"`
	Facs    string `xml:"facs,attr,omitempty"`
	Corresp string `xml:"corresp,attr,omitempty"`

	Nodes []ABNode `xml:"-"` // mixed content (anchors, text, inline, etc.)
}

// Anchor is a TEI anchor (empty element with xml:id) for stand-off from/to.
type Anchor struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
}

// ABNode is one item in mixed content: chardata, line break, anchor, or inline (e.g. name).
type ABNode struct {
	CharData string  `xml:"-"` // emitted as chardata when LB, Inline, Anchor are nil
	Anchor   *Anchor `xml:"-"`
	Inline   *Inline `xml:"-"`
}

// AB is a block of text (TEI ab). Use either Segs (one seg per line) or Nodes (mixed content with lb).
type AB struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	Facs  string `xml:"facs,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`

	Lines []L `xml:"l,omitempty"`
}

// MarshalXML implements xml.Marshaler for <l> with mixed content nodes.
func (l L) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "l"}

	var attrs []xml.Attr
	if l.XmlID != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "xml:id"}, Value: l.XmlID})
	}
	if l.Facs != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "facs"}, Value: l.Facs})
	}
	if l.Corresp != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "corresp"}, Value: l.Corresp})
	}
	start.Attr = attrs

	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for i := range l.Nodes {
		if err := l.Nodes[i].MarshalXML(e); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}
