package formatcov

// COCO types

type cocoImage struct {
	ID       int    `json:"id"`
	FileName string `json:"file_name"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	File     string `json:"file,omitempty"`
}

type cocoAnnotation struct {
	ID           int         `json:"id"`
	ImageID      int         `json:"image_id"`
	CategoryID   int         `json:"category_id"`
	BBox         []float64   `json:"bbox"` // [x_min, y_min, width, height]
	Area         float64     `json:"area"`
	IsCrowd      int         `json:"iscrowd"`
	Score        float64     `json:"score,omitempty"`
	Segmentation [][]float64 `json:"segmentation,omitempty"`
}

type cocoCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type cocoInfo struct {
	Description string `json:"description"`
	URL         string `json:"url"`
	Version     string `json:"version"`
	Year        int    `json:"year"`
	Contributor string `json:"contributor"`
	DateCreated string `json:"date_created"`
}

type cocoLicense struct {
	URL  string `json:"url"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type cocoRoot struct {
	Images      []cocoImage      `json:"images"`
	Annotations []cocoAnnotation `json:"annotations"`
	Categories  []cocoCategory   `json:"categories"`
	Info        cocoInfo         `json:"info"`
	Licenses    []cocoLicense    `json:"licenses"`
}
