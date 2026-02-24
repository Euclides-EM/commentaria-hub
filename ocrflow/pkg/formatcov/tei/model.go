package tei

// tei.Xmlns = "http://www.tei-c.org/ns/1.0"

type LinesInput struct {
	LinesByKeys map[string]Lines
}

type Lines struct {
	TranscriptionLines []string
	Translations       map[string][]string
}

type EntitiesInput struct {
	// Occurrences to mark in the transcription.
	// Each occurrence maps to exactly one ALTO TextLine by (PageID, BlockID, LineID).
	Occurrences []EntityOccurrence
	// Profiles to include in the TEI header. Each profile has an ID that can be referenced by occurrences.
	Profiles map[string][]Profile
}

type EntityOccurrence struct {
	PageID  string // ALTO Page.ID
	BlockID string // ALTO TextBlock.ID
	LineID  string // ALTO TextLine.ID

	// Start and End are byte offsets in the joined line text.
	// End is exclusive.
	Start int
	End   int

	// TEI element to emit: "persName", "orgName", "placeName", "name", etc.
	Element string

	// Link to entity profile in header, e.g. "#ent_ibn_rushd"
	Ref string

	// Optional pointer to feature taxonomy, e.g. "#feat_person"
	Ana string
}

type Profile struct {
	Type  string // "latinized_name", "original_name", "birth_date", "death_date", etc.
	Value string
}
