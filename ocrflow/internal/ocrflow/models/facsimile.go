package models

type Facsimile struct {
	ID           string `json:"id"`
	ScanURL      string `json:"scan_url"`
	PDFLocalPath string `json:"local_path"`
	JPGsLocalDir string `json:"jpgs_local_dir"`
}

func (f Facsimile) DeepCopy() *Facsimile {
	return &Facsimile{
		ID:           f.ID,
		ScanURL:      f.ScanURL,
		PDFLocalPath: f.PDFLocalPath,
		JPGsLocalDir: f.JPGsLocalDir,
	}
}
