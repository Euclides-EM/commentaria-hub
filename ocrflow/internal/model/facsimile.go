package model

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type Facsimile struct {
	common.Meta
	EditionID     string `json:"edition_id"`
	ScanURL       string `json:"scan_url"`
	MainTextPages string `json:"main_text_pages"`
}
