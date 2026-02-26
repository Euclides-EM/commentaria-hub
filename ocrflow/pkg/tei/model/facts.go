package model

// StandOff holds stand-off annotations in teiHeader (mentions spanGrp, relations, interpretation groups).
type StandOff struct {
	SpanGrp    *SpanGrp    `xml:"spanGrp,omitempty"`
	InterpGrps []InterpGrp `xml:"interpGrp,omitempty"`
}

// SpanGrp holds a layer of span elements (e.g. ner-mentions) referencing anchors via from/to.
type SpanGrp struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`
	Spans []Span `xml:"span,omitempty"`
}

// Span is a single stand-off span (from/to anchor refs, ana, ref to entity).
type Span struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	From  string `xml:"from,attr,omitempty"`
	To    string `xml:"to,attr,omitempty"`
	Ana   string `xml:"ana,attr,omitempty"`

	Notes []Note `xml:"note,omitempty"`
}

// InterpGrp is a generic fallback group of interpretations (TEI interpGrp).
type InterpGrp struct {
	XmlID   string   `xml:"xml:id,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
	Interps []Interp `xml:"interp,omitempty"`
}

// Interp is a single interpretation (TEI interp).
type Interp struct {
	XmlID string `xml:"xml:id,attr,omitempty"`
	Text  string `xml:",chardata"`
}
