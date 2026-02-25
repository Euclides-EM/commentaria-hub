package model

import "encoding/xml"

// LB is a line break (TEI lb); used in mixed-content transcription.
type LB struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	Facs  string `xml:"facs,attr,omitempty"`
}

// Anchor is a TEI anchor (empty element with xml:id) for stand-off from/to.
type Anchor struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
}

// ABNode is one item in mixed content: chardata, line break, anchor, or inline (e.g. name).
type ABNode struct {
	CharData string  `xml:"-"` // emitted as chardata when LB, Inline, Anchor are nil
	LB       *LB     `xml:"-"`
	Anchor   *Anchor `xml:"-"`
	Inline   *Inline `xml:"-"`
}

// AB is a block of text (TEI ab). Use either Segs (one seg per line) or Nodes (mixed content with lb).
type AB struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	Facs  string `xml:"facs,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`

	Segs  []Seg    `xml:"seg,omitempty"`
	Nodes []ABNode `xml:"-"` // when set, encoded as mixed content (no wrapper element)
}

// MarshalXML implements xml.Marshaler. When Nodes is non-empty, emits mixed content (chardata, lb, inline); otherwise uses default encoding for Segs.
func (a AB) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "ab"}
	var attrs []xml.Attr
	if a.XmlID != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "xml:id"}, Value: a.XmlID})
	}
	if a.Facs != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "facs"}, Value: a.Facs})
	}
	if a.Type != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "type"}, Value: a.Type})
	}
	start.Attr = attrs

	if len(a.Nodes) > 0 {
		if err := e.EncodeToken(start); err != nil {
			return err
		}
		for i := range a.Nodes {
			if err := a.Nodes[i].MarshalXML(e); err != nil {
				return err
			}
		}
		return e.EncodeToken(start.End())
	}
	// Default: encode as regular struct (Segs). Use minimal start so attributes
	// come only from the struct; passing start with attrs would duplicate them.
	return e.EncodeElement(struct {
		XmlID string `xml:"xml:id,attr,omitempty"`
		Facs  string `xml:"facs,attr,omitempty"`
		Type  string `xml:"type,attr,omitempty"`
		Segs  []Seg  `xml:"seg,omitempty"`
	}{
		XmlID: a.XmlID,
		Facs:  a.Facs,
		Type:  a.Type,
		Segs:  a.Segs,
	}, xml.StartElement{Name: xml.Name{Local: "ab"}})
}

// MarshalXML implements xml.Marshaler for one item of mixed content (chardata, lb, anchor, or inline).
func (n ABNode) MarshalXML(e *xml.Encoder) error {
	if n.LB != nil {
		start := xml.StartElement{Name: xml.Name{Local: "lb"}}
		var attrs []xml.Attr
		if n.LB.XmlID != "" {
			attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "xml:id"}, Value: n.LB.XmlID})
		}
		if n.LB.Facs != "" {
			attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "facs"}, Value: n.LB.Facs})
		}
		start.Attr = attrs
		if err := e.EncodeToken(start); err != nil {
			return err
		}
		return e.EncodeToken(start.End())
	}
	if n.Anchor != nil && n.Anchor.XmlID != "" {
		start := xml.StartElement{Name: xml.Name{Local: "anchor"}}
		start.Attr = []xml.Attr{{Name: xml.Name{Local: "xml:id"}, Value: n.Anchor.XmlID}}
		if err := e.EncodeToken(start); err != nil {
			return err
		}
		return e.EncodeToken(start.End())
	}
	if n.Inline != nil {
		start := xml.StartElement{Name: xml.Name{Local: n.Inline.XMLName.Local}}
		if n.Inline.XMLName.Local == "" {
			start.Name = xml.Name{Local: "name"}
		}
		return n.Inline.MarshalXML(e, start)
	}
	if n.CharData != "" {
		return e.EncodeToken(xml.CharData(n.CharData))
	}
	return nil
}
