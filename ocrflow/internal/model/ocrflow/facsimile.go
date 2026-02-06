package ocrflow

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type Facsimile struct {
	common.Meta
	ScanURL       string `json:"scan_url"`
	MainTextPages string `json:"main_text_pages"`
}
