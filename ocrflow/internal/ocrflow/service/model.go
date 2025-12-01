package service

import (
	"errors"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"path"
)

type Model struct {
	m         map[string]*model.Model
	modelsDir string
}

func NewModelService(modelsDir string) *Model {
	return &Model{
		m: map[string]*model.Model{
			"paris1615trained_2811": {
				Meta:      model.NewMeta("paris1615trained_2811"),
				LocalPath: path.Join(modelsDir, "paris1615trained_2811.pt"),
				Type:      model.OCRModelTypeOCR,
				Location:  model.OCRModelLocationLocal,
				Name:      "paris1615trained_2811",
			},
			"CapricciosaM": {
				Meta:      model.NewMeta("CapricciosaM"),
				LocalPath: path.Join(modelsDir, "CapricciosaM.pt"),
				Type:      model.OCRModelTypeSegment,
				Location:  model.OCRModelLocationLocal,
				Name:      "CapricciosaM",
			},
			"Gallicorpor": {
				Meta:      model.NewMeta("Gallicorpor"),
				LocalPath: path.Join(modelsDir, "Gallicorpor.mlmodel"),
				Type:      model.OCRModelTypeOCR,
				Location:  model.OCRModelLocationLocal,
				Name:      "Gallicorpor",
				// todo: add categories
			},
			"Paris1615NoContinuedPNoMainZone3": {
				Meta:            model.NewMeta("Paris1615NoContinuedPNoMainZone3"),
				Type:            model.OCRModelTypeSegment,
				Location:        model.OCRModelLocationRoboflow,
				AlgorithmFamily: model.OCRModelAlgorithmFamilyYOLO,
				Name:            "paris-1615-nocontinuedpnomainzone-dbxgq/3",
			},
			"Paris1615Polygons1": {
				Meta:            model.NewMeta("Paris1615Polygons1"),
				Type:            model.OCRModelTypeSegment,
				Location:        model.OCRModelLocationRoboflow,
				AlgorithmFamily: model.OCRModelAlgorithmFamilyYOLO,
				Name:            "paris-1615-polygons-h4cad/1",
			},
			"Paris1615PolygonsAndMainZone": {
				Meta:            model.NewMeta("Paris1615PolygonsAndMainZone"),
				Type:            model.OCRModelTypeSegment,
				Location:        model.OCRModelLocationRoboflow,
				AlgorithmFamily: model.OCRModelAlgorithmFamilyYOLO,
				Name:            "paris-1615-polygonswithmz-wsrge/1",
			},
			"Paris1615NoMainZoneSubtypes": {
				Meta:            model.NewMeta("Paris1615NoMainZoneSubtypes"),
				Type:            model.OCRModelTypeSegment,
				Location:        model.OCRModelLocationRoboflow,
				AlgorithmFamily: model.OCRModelAlgorithmFamilyYOLO,
				Name:            "paris-1615-withmznosubtypes-tkgii/1",
			},
			"segmontoRB": {
				Meta:            model.NewMeta("segmontoRB"),
				Type:            model.OCRModelTypeSegment,
				Location:        model.OCRModelLocationRoboflow,
				AlgorithmFamily: model.OCRModelAlgorithmFamilyYOLO,
				Name:            "segmonto/31",
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

func (m *Model) Upsert(mo *model.Model, modelPath string) error {
	n := mo.DeepCopy()
	n.ID = idgen.GenerateID()
	n.Location = model.OCRModelLocationLocal
	n.LocalPath = path.Join(m.modelsDir, fmt.Sprintf("%s.pt", n.ID))
	if err := futils.CopyFile(modelPath, n.LocalPath); err != nil {
		return fmt.Errorf("failed to copy model from %s to %s: %w", modelPath, n.LocalPath, err)
	}
	m.m[n.ID] = n.DeepCopy()
	return nil
}
