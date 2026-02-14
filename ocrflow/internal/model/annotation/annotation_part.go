package annotation

type Part struct {
	Category string   `json:"category,omitempty"`
	Content  string   `json:"content,omitempty"`
	Location Location `json:"location"`
}

type Location struct {
	Page        int    `json:"page"`
	TextBlockID string `json:"text_block_id,omitempty"`
}
