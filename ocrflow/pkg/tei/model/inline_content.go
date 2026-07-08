package model

import "encoding/xml"

type Inline struct {
	XMLName xml.Name

	Text string `xml:",chardata"`

	XmlID   string `xml:"xml:id,attr,omitempty"`
	Corresp string `xml:"corresp,attr,omitempty"`
	Facs    string `xml:"facs,attr,omitempty"`
	Ref     string `xml:"ref,attr,omitempty"`
	Target  string `xml:"target,attr,omitempty"`
	Ana     string `xml:"ana,attr,omitempty"`
	Rend    string `xml:"rend,attr,omitempty"`
}

// MarshalXML implements xml.Marshaler so that plain text (XMLName.Local == "")
// is emitted as chardata only, and entity inlines as proper TEI elements.
func (i Inline) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if i.XMLName.Local == "" {
		return e.EncodeToken(xml.CharData(i.Text))
	}
	start.Name = i.XMLName
	var attrs []xml.Attr
	if i.XmlID != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "xml:id"}, Value: i.XmlID})
	}
	if i.Corresp != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "corresp"}, Value: i.Corresp})
	}
	if i.Facs != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "facs"}, Value: i.Facs})
	}
	if i.Ref != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "ref"}, Value: i.Ref})
	}
	if i.Target != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "target"}, Value: i.Target})
	}
	if i.Ana != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "ana"}, Value: i.Ana})
	}
	if i.Rend != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "rend"}, Value: i.Rend})
	}
	start.Attr = attrs
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if i.Text != "" {
		if err := e.EncodeToken(xml.CharData(i.Text)); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}
