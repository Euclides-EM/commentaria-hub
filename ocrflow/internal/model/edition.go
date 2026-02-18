package model

type Edition struct {
	Key                   string                 `json:"key"`
	ShortTitle            string                 `json:"shortTitle"`
	ShortTitleSource      string                 `json:"shortTitleSource"`
	Notes                 string                 `json:"notes"`
	Corpus                []string               `json:"corpus"`
	Shelfmarks            []EditionShelfmark     `json:"shelfmarks"`
	Verified              bool                   `json:"verified"`
	Bibliography          []string               `json:"bibliography"`
	ReprintOf             *string                `json:"reprintOf"`
	VisualElements        []EditionVisualElement `json:"visualElements"`
	DiagramCropsAvailable bool                   `json:"diagramCropsAvailable"`
	HasDiagrams           bool                   `json:"HasDiagrams"`

	// Manuscript-only
	IsManuscript       bool    `json:"isManuscript"`
	ManuscriptYearFrom *int    `json:"manuscriptYearFrom"`
	ManuscriptYearTo   *int    `json:"manuscriptYearTo"`
	ManuscriptClass    string  `json:"manuscriptClass"`
	ManuscriptSubclass *string `json:"manuscriptSubclass"`

	// Print-only
	Cities         []string `json:"cities"`
	Year           *string  `json:"year"`
	Languages      []string `json:"languages"`
	Editor         []string `json:"editor"`
	Publisher      []string `json:"publisher"`
	Format         *int     `json:"format"`
	Volumes        *int     `json:"volumes"`
	USTCId         *string  `json:"ustcId"`
	Title          *string  `json:"title"`
	TitleEN        *string  `json:"title_EN"`
	Imprint        *string  `json:"imprint"`
	ImprintEN      *string  `json:"imprint_EN"`
	Colophon       *string  `json:"colophon"`
	ColophonEN     *string  `json:"colophon_EN"`
	Frontispiece   *string  `json:"frontispiece"`
	FrontispieceEN *string  `json:"frontispiece_EN"`

	// Elements (both)
	IsElements        bool     `json:"isElements"`
	Books             []int    `json:"books"`
	AdditionalContent []string `json:"additionalContent"`
}

type EditionShelfmark struct {
	Volume          *int   `json:"volume"`
	Scan            string `json:"scan"`
	Shelfmark       string `json:"shelfmark"`
	TitlePageImg    string `json:"title_page_img"`
	FrontispieceImg string `json:"frontispiece_img"`
	Annotations     string `json:"annotations"`
	Copyright       string `json:"copyright"`
}

type EditionVisualElement struct {
	VisualElementType string                 `json:"visual_element_type"`
	LocatorType       string                 `json:"locator_type"`
	Notes             string                 `json:"notes"`
	Locator           *EditionLocator        `json:"locator"`
	Examples          []EditionVisualExample `json:"examples"`
}

type EditionLocator struct {
	Key             string  `json:"key"`
	FirstOrderType  *string `json:"first_order_type"`
	FirstOrderValue *string `json:"first_order_value"`
	Type            *string `json:"type"`
	Value           string  `json:"value"`
	PageType        string  `json:"page_type"`
	PageValue       *string `json:"page_value"`
}

type EditionVisualExample struct {
	Img        string          `json:"img"`
	HasLocator bool            `json:"has_locator"`
	Locator    *EditionLocator `json:"locator"`
}

// EditionListResult is the paginated response for listing editions.
type EditionListResult struct {
	Items  []*Edition `json:"items"`
	Total  int        `json:"total"`
	Offset int        `json:"offset"`
	Limit  int        `json:"limit"`
}
