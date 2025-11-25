package model

type Facsimile struct {
	Meta
	ScanURL string `json:"scan_url"`
}

func (f Facsimile) DeepCopy() *Facsimile {
	return &Facsimile{
		Meta:    f.Meta.DeepCopy(),
		ScanURL: f.ScanURL,
	}
}
