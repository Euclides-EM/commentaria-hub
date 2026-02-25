package model

// StandOff holds stand-off annotations in teiHeader (mentions spanGrp, relations, interpretation groups).
type StandOff struct {
	SpanGrp      *SpanGrp      `xml:"spanGrp,omitempty"`
	ListRelation *ListRelation `xml:"listRelation,omitempty"`
	InterpGrps   []InterpGrp   `xml:"interpGrp,omitempty"`
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
	Ref   string `xml:"ref,attr,omitempty"`
}

// ListRelation wraps a list of relation elements (TEI listRelation).
type ListRelation struct {
	Relations []Relation `xml:"relation,omitempty"`
}

// Relation represents a typed relation between two entities (TEI relation).
// For feature-assignment facts: passive=entity, corresp=mention(s) as evidence.
type Relation struct {
	XmlID   string `xml:"xml:id,attr,omitempty"`
	Name    string `xml:"name,attr"`
	Active  string `xml:"active,attr,omitempty"`
	Passive string `xml:"passive,attr,omitempty"`
	Ana     string `xml:"ana,attr,omitempty"`
	Cert    string `xml:"cert,attr,omitempty"`
	Source  string `xml:"source,attr,omitempty"`
	Corresp string `xml:"corresp,attr,omitempty"`
}

// InterpGrp is a generic fallback group of interpretations (TEI interpGrp).
type InterpGrp struct {
	Type    string   `xml:"type,attr,omitempty"`
	Interps []Interp `xml:"interp,omitempty"`
}

// Interp is a single interpretation (TEI interp).
type Interp struct {
	XmlID   string `xml:"xml:id,attr,omitempty"`
	Type    string `xml:"type,attr,omitempty"`
	Corresp string `xml:"corresp,attr,omitempty"`
	Text    string `xml:",chardata"`
}
