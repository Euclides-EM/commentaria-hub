package ocrflow

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type Dataset struct {
	common.Meta
	FacsimileID string  `json:"facsimile_id"`
	EditionID   string  `json:"edition_id"`
	DPI         float64 `json:"dpi" default:"300"`
	Deskewed    bool    `json:"deskewed" default:"false"`
}
