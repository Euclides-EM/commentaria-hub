package service

import (
	"errors"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"path"
)

type Model struct {
	m         map[string]*model.Model
	modelsDir string
}

func NewModelService(modelsDir string) *Model {
	return &Model{
		m: map[string]*model.Model{
			"CapricciosaM": {
				Meta:      model.NewMeta("CapricciosaM"),
				LocalPath: path.Join(modelsDir, "CapricciosaM.pt"),
				Type:      model.OCRModelTypeSegment,
				RunWith:   "kraken",
				Name:      "CapricciosaM",
			},
			"Gallicorpor": {
				Meta:      model.NewMeta("Gallicorpor"),
				LocalPath: path.Join(modelsDir, "Gallicorpor.mlmodel"),
				Type:      model.OCRModelTypeOCR,
				RunWith:   "kraken",
				Name:      "Gallicorpor",
				// todo: add categories
			},
			"segmontoRB": {
				Meta:    model.NewMeta("segmontoRB"),
				Type:    model.OCRModelTypeSegment,
				RunWith: "roboflow",
				Name:    "segmonto/31",
				Categories: []string{
					"AdvertisementZone",
					"DigitizationArtefactZone",
					"DropCapitalZone",
					"FigureZone",
					"FigureZone-FigDesc",
					"FigureZone-Head",
					"FormZone",
					"GraphicZone",
					"GraphicZone-Decoration",
					"GraphicZone-FigDesc",
					"GraphicZone-Head",
					"GraphicZone-Maths",
					"GraphicZone-Part",
					"GraphicZone-TextualContent",
					"MainZone-Continued",
					"MainZone-Date",
					"MainZone-Entry",
					"MainZone-Entry-Continued",
					"MainZone-Head",
					"MainZone-Lg",
					"MainZone-Lg-Continued",
					"MainZone-List-Continued",
					"MainZone-ListItem",
					"MainZone-Maths",
					"MainZone-Other",
					"MainZone-P",
					"MainZone-P-Continued",
					"MainZone-Signature",
					"MainZone-Sp",
					"MainZone-Sp-Continued",
					"MarginTextZone-ContinuedNotes",
					"MarginTextZone-ManuscriptAddendum",
					"MarginTextZone-Notes",
					"MarginTextZone-Notes-Continued",
					"MusicZone",
					"NumberingZone",
					"PageTitleZone",
					"PageTitleZone-Index",
					"QuireMarksZone",
					"RunningTitleZone",
					"StampZone",
					"StampZone-Sticker",
					"TableZone",
					"TableZone-Continued",
					"TableZone-Head",
					"TitlePageZone",
					"TitlePageZone-Index",
				},
			},
		},
	}
}

func (m *Model) List() ([]*model.Model, error) {
	modelsList := make([]*model.Model, 0, len(m.m))
	for _, retrieved := range m.m {
		modelsList = append(modelsList, retrieved.DeepCopy())
	}
	return modelsList, nil
}

func (m *Model) Get(id string) (*model.Model, error) {
	retrieved, ok := m.m[id]
	if !ok {
		return nil, errors.New("model not found")
	}
	return retrieved.DeepCopy(), nil
}
