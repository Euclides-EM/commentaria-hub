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
	PageID  string
	BlockID string
	LineID  string
	Start   int // byte offset, inclusive
	End     int // byte offset, exclusive

	Element string // e.g. "persName", "orgName"
	Ana     string

	// Profile / relation (optional). When Value is set, contributes to teiHeader profileDesc (keywords).
	// When ObjectRef is set, contributes to standOff listRelation.
	Type               string
	Value              string
	ObjectRef          string
	Cert               float64
	EvidenceMentionIDs []string
}
