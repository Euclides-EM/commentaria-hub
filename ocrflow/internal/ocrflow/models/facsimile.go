package models

type Facsimile struct {
	ID        string `json:"id"`
	ScanURL   string `json:"scan_url"`
	LocalPath string `json:"local_path"`
}
