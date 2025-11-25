package model

type Dataset struct {
	Meta
	Facsimile  Reference `json:"facsimile"`
	Edition    Reference `json:"edition"`
	PDFPath    string    `json:"pdf_path" readonly:"true"`
	ImagesPath string    `json:"img_path" readonly:"true"`
}

func (d *Dataset) FacsimileID() string {
	return d.Facsimile.ID
}

func (d *Dataset) EditionID() string {
	return d.Edition.ID
}
