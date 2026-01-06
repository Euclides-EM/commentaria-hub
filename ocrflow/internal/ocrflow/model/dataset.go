package model

type Dataset struct {
	Meta
	FacsimileID string  `json:"facsimile_id"`
	EditionID   string  `json:"edition_id"`
	PDFPath     string  `json:"pdf_path" readonly:"true"`
	ImagesPath  string  `json:"img_path" readonly:"true"`
	DPI         float64 `json:"dpi" default:"300"`
}
