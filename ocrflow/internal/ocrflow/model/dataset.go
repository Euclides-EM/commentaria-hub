package model

type Dataset struct {
	Meta
	FacsimileID string  `json:"facsimile_id"`
	EditionID   string  `json:"edition_id"`
	DPI         float64 `json:"dpi" default:"300"`
	Deskewed    bool    `json:"deskewed" default:"false"`
}
