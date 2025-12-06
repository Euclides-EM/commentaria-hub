package service

import (
	"errors"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/tiendc/go-deepcopy"
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
			// Fine-tuned model based on:
			// The CapricciosaM model, further fine-tuned on pages from the London 1570 facsimile
			// Annotations in https://app.roboflow.com/mia-workplace/1570-english/2
			"1570FineTuned_0312": {
				Meta:      model.NewMeta("1570FineTuned_0312"),
				LocalPath: path.Join(modelsDir, "1570FineTunedCapricciosaM_0312.pt"),
				Type:      model.OCRModelTypeSegment,
				Location:  model.OCRModelLocationLocal,
				Name:      "CapricciosaM",
			},
			// Fine-tuned model based on:
			// The CapricciosaM model, further fine-tuned on pages from the Paris 1615 facsimile
			// Annotations in https://app.roboflow.com/mia-workplace/0212-xcfg/2
			"1615FineTunedCapricciosaM_0312": {
				Meta:      model.NewMeta("1615FineTunedCapricciosaM_0312"),
				LocalPath: path.Join(modelsDir, "1615FineTunedCapricciosaM_0312.pt"),
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
			// With a main zone, subtypes, based on the data set after skewing
			"0212-xcfg/2": {
				Meta:            model.NewMeta("0212-xcfg-2"),
				Type:            model.OCRModelTypeSegment,
				Location:        model.OCRModelLocationRoboflow,
				AlgorithmFamily: model.OCRModelAlgorithmFamilyYOLO,
				Name:            "0212-xcfg/2",
			},
			// Model trained on London 1570 facsimile, with main zone and heading (no additional subtypes)
			// Can be found here: https://app.roboflow.com/mia-workplace/1570-english/models/1570-english/2
			"1570-english/2": {
				Meta:            model.NewMeta("1570-english-2"),
				Type:            model.OCRModelTypeSegment,
				Location:        model.OCRModelLocationRoboflow,
				AlgorithmFamily: model.OCRModelAlgorithmFamilyYOLO,
				Name:            "1570-english/2",
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
		var dst *model.Model
		if err := deepcopy.Copy(&dst, &retrieved); err != nil {
			return nil, fmt.Errorf("failed to copy annotation: %w", err)
		}
		modelsList = append(modelsList, dst)
	}
	return modelsList, nil
}

func (m *Model) Get(id string) (*model.Model, error) {
	retrieved, ok := m.m[id]
	if !ok {
		return nil, errors.New("model not found")
	}
	var dst *model.Model
	if err := deepcopy.Copy(&dst, &retrieved); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	return dst, nil
}

func (m *Model) Upsert(mo *model.Model, modelPath string) error {
	var n *model.Model
	if err := deepcopy.Copy(&n, &mo); err != nil {
		return fmt.Errorf("failed to copy annotation: %w", err)
	}
	n.ID = idgen.GenerateID()
	n.Location = model.OCRModelLocationLocal
	n.LocalPath = path.Join(m.modelsDir, fmt.Sprintf("%s.pt", n.ID))
	if err := futils.CopyFile(modelPath, n.LocalPath); err != nil {
		return fmt.Errorf("failed to copy model from %s to %s: %w", modelPath, n.LocalPath, err)
	}
	var dst *model.Model
	if err := deepcopy.Copy(&dst, &n); err != nil {
		return fmt.Errorf("failed to copy annotation: %w", err)
	}
	m.m[n.ID] = dst
	return nil
}
