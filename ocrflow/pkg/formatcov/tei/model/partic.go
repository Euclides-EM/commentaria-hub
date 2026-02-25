package model

// ParticDesc holds participant/entity descriptions (listOrg and list).
type ParticDesc struct {
	List *List `xml:"list,omitempty"`
}

// List holds item elements for generic entities.
type List struct {
	Items []Item `xml:"item,omitempty"`
}

// Item is a single entity in list (label + notes).
type Item struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	Label string `xml:"label,omitempty"`
	Notes []Note `xml:"note,omitempty"`
}

// Note is a TEI note (e.g. type="feature" ana="#feat_*", or type="normalized").
type Note struct {
	Type string `xml:"type,attr,omitempty"`
	Ana  string `xml:"ana,attr,omitempty"`
	Text string `xml:",chardata"`
}
