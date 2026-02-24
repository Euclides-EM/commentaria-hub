package common

type ALTOPart struct {
	Category string       `json:"category,omitempty"`
	Content  string       `json:"content,omitempty"`
	Location ALTOLocation `json:"location"`
}

type ALTOLocation struct {
	Page           int    `json:"page"`
	TextBlockID    string `json:"text_block_id,omitempty"`
	TextLineID     string `json:"text_line_id,omitempty"`
	CharactersSpan *Span  `json:"characters_span,omitempty"`
}
