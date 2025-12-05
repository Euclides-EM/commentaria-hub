package model

type Facsimile struct {
	Meta
	ScanURL string `json:"scan_url"`
}
