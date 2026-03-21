package common

type ALTOPart struct {
	Category string       `json:"category,omitempty"`
	Content  string       `json:"content,omitempty"`
	Location ALTOLocation `json:"location"`
}

type ALTOLocation struct {
	Page        string `json:"page"`
	TextBlockID string `json:"text_block_id,omitempty"`
}
