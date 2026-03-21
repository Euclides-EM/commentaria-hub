package annotation

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type Search struct {
	SearchWithin []SearchWithin     `json:"search_within"`
	Categories   []string           `json:"categories"`
	Regex        string             `json:"regex"`
	DatasetID    string             `json:"dataset_id"`
	AnnotationId string             `json:"annotation_id"`
	MaxResults   int                `json:"max_results"`
	Results      []*common.ALTOPart `json:"results" readonly:"true"`
}

type SearchWithin string

const (
	SearchWithinCategories     SearchWithin = "categories"
	SearchWithinTranscription  SearchWithin = "transcription"
	SearchWithinTranslation    SearchWithin = "translation"
	SearchWithinBiblioMetadata SearchWithin = "biblio_metadata"
)

func ToSearchWithin(vals []string) []SearchWithin {
	var res []SearchWithin
	for _, v := range vals {
		switch v {
		case string(SearchWithinCategories):
			res = append(res, SearchWithinCategories)
		case string(SearchWithinTranscription):
			res = append(res, SearchWithinTranscription)
		case string(SearchWithinTranslation):
			res = append(res, SearchWithinTranslation)
		case string(SearchWithinBiblioMetadata):
			res = append(res, SearchWithinBiblioMetadata)
		}
	}
	return res
}
