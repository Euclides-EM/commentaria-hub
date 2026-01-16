package model

type Facsimile struct {
	Meta
	ScanURL       string `json:"scan_url"`
	MainTextPages string `json:"main_text_pages"`
}
