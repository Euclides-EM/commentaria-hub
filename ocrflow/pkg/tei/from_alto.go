package tei

import (
	"encoding/xml"
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
	"github.com/samber/lo"
)

func BuildTEIFromALTO(
	pageKey string,
	a *alto.Alto,
	entities []EntityItem,
	imageUrl string,
) (*model.TEI, error) {

	if len(a.Layout.Page) != 1 {
		return nil, fmt.Errorf("expected exactly one page in ALTO, got %d", len(a.Layout.Page))
	}

	doc := &model.TEI{
		XMLName: xml.Name{Local: "TEI"},
		Xmlns:   "http://www.tei-c.org/ns/1.0",
		Header: model.Header{
			FileDesc: buildFileDesc(),
			StandOff: buildStandOff(entities),
		},
		Facsimile: buildFacsimileForAlto(pageKey, imageUrl, a),
		Text:      model.Text{Body: model.Body{}},
	}

	doc.Text.Body.PBs = append(doc.Text.Body.PBs, model.PB{
		Facs: "#" + doc.Facsimile.Surfaces[0].XmlID,
		N:    pageKey,
	})

	var abs []model.AB
	for i, textBlock := range a.Layout.Page[0].PrintSpace.TextBlocks {
		ab := model.AB{
			XmlID: transcriptionAnonBlockID(pageKey, i+1),
			Type:  transcriptionAnonBlockType,
			Facs:  "#" + facZoneBlockID(pageKey, i+1),
		}
		for j, textLine := range textBlock.Lines {
			entitiesForLine := lo.Filter(entities, func(e EntityItem, _ int) bool {
				return (e.Start.LineID == textLine.ID && e.Start.BlockID == textBlock.ID) || (e.End.LineID == textLine.ID && e.End.BlockID == textBlock.ID)
			})
			nodes := buildInlineNodesWithAnchors(textBlock.ID, textLine.ID, alto.ExtractTextFromLine(textLine), entitiesForLine)
			l := model.L{
				XmlID: lineID(pageKey, i+1, j+1),
				Facs:  "#" + facZoneLineID(pageKey, i+1, j+1),
				Nodes: nodes,
			}
			ab.Lines = append(ab.Lines, l)
		}
		abs = append(abs, ab)
	}

	transDiv := model.Div{
		Type: "transcription",
		N:    pageKey,
		Abs:  abs,
	}
	doc.Text.Body.Divs = append(doc.Text.Body.Divs, transDiv)

	return doc, nil
}
