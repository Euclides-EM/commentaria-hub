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
	Ref string // entity ref; normalized to # form when building

	// Inline mention (optional). When set, this item is emitted in the text.
	Start EntityLocationIndex // byte offset is inclusive
	End   EntityLocationIndex // byte offset is exclusive

	Ana string

	// Profile / relation (optional). When Value is set, contributes to teiHeader profileDesc (keywords).
	Type               string
	Value              string
	Cert               float64
	EvidenceMentionIDs []string
}

type EntityLocationIndex struct {
	PageID     string
	BlockID    string
	LineID     string
	ByteOffset int
}
