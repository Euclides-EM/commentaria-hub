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

	Category   string // same as "feature"; used for standoff ana
	Properties map[string]string
}

type EntityLocationIndex struct {
	BlockID    string
	LineID     string
	ByteOffset int
}
