package tei

import (
	"fmt"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

func buildFacsimileForLines(pageKey, imageUrl string, lines []string) model.Facsimile {
	textBlockZones := []model.Zone{
		{
			XmlID: facZoneBlockID(pageKey, 1),
			Type:  textBlockZoneType,
		},
	}
	var lineZones []model.Zone
	for i := range lines {
		lineZoneID := facZoneLineID(pageKey, 1, i+1)
		lineZones = append(lineZones, model.Zone{
			XmlID: lineZoneID,
			Type:  textLineZoneType,
		})
	}

	return model.Facsimile{
		Surfaces: []model.Surface{
			{
				XmlID: surfaceID(pageKey),
				Facs:  imageUrl,
				N:     pageKey,
				Zones: append(textBlockZones, lineZones...),
			},
		},
	}
}

func buildFacsimileForAlto(pageKey, imageUrl string, a *alto.Alto) model.Facsimile {
	var textBlockZones []model.Zone
	var lineZones []model.Zone
	for i, textBlock := range a.Layout.Page[0].PrintSpace.TextBlocks {
		textBlockZoneID := facZoneBlockID(pageKey, i+1)
		textBlockZones = append(textBlockZones, model.Zone{
			XmlID:   textBlockZoneID,
			Type:    textBlockZoneType,
			Ana:     altoCategoryAna(textBlock.TagRefs, a.Tags),
			ULX:     textBlock.HPOS,
			ULY:     textBlock.VPOS,
			LRX:     textBlock.HPOS + textBlock.Width,
			LRY:     textBlock.VPOS + textBlock.Height,
			Corresp: fmt.Sprintf("alto:textblock:%s", textBlock.ID),
		})
		for j, tb := range textBlock.Lines {
			lineZoneID := facZoneLineID(pageKey, i+1, j+1)
			lineZones = append(lineZones, model.Zone{
				XmlID:   lineZoneID,
				Type:    textLineZoneType,
				ULX:     tb.HPOS,
				ULY:     tb.VPOS,
				LRX:     tb.HPOS + tb.Width,
				LRY:     tb.VPOS + tb.Height,
				Corresp: fmt.Sprintf("alto:textline:%s", tb.ID),
			})
		}
	}

	return model.Facsimile{
		Surfaces: []model.Surface{
			{
				XmlID: surfaceID(pageKey),
				Facs:  imageUrl,
				N:     pageKey,
				LRX:   float64(a.Layout.Page[0].Width),
				LRY:   float64(a.Layout.Page[0].Height),
				Zones: append(textBlockZones, lineZones...),
			},
		},
	}
}
