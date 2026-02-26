package tei

// tei.Xmlns = "http://www.tei-c.org/ns/1.0"

type LinesInput struct {
	LinesByKeys map[string]Lines
}

type Lines struct {
	TranscriptionLines []string
	Translations       map[string][]string
}

// EntityItem is a single mention and/or profile/relation row. Ref identifies the entity (e.g. "ent_john" or "#ent_john").
type EntityItem struct {
	Start EntityLocationIndex // byte offset is inclusive
	End   EntityLocationIndex // byte offset is exclusive

	Ref   string  // entity reference (e.g. "ent_john")
	Ana   string  // interpretation, e.g. "#feat_person" (used as Category when Category is empty)
	Type  string  // fact/relation type
	Value string  // fact value
	Cert  float64 // certainty 0–1

	Category   string // same as "feature"; used for standoff ana
	Properties map[string]string
}

type EntityLocationIndex struct {
	PageID     string // optional, for test data
	BlockID    string
	LineID     string
	ByteOffset int
}
