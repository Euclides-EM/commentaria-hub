package tei

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
)

func BuildTEIFromALTO(
	a *alto.Alto,
	entities []EntityItem,
	imageUrl string,
) (*model.TEI, error) {

	if a == nil {
		return nil, fmt.Errorf("alto is nil")
	}

	const entityBlockID = "b1" // for entity targeting we use a fixed block id
	lineTextByKey := make(map[string]string)
	orderedKeys := make([]string, 0)
	for _, pg := range a.Layout.Page {
		for _, blk := range pg.PrintSpace.TextBlocks {
			for i, ln := range blk.Lines {
				k := occKey(pg.ID, entityBlockID, fmt.Sprintf("l%04d", i+1))
				orderedKeys = append(orderedKeys, k)
				lineTextByKey[k] = joinLineStrings(ln.Strings)
			}
		}
	}
	getLineText := func(key string) string { return lineTextByKey[key] }
	occByKey := buildEntitiesOccurrences(entities, orderedKeys, getLineText)

	doc := &model.TEI{
		XMLName: xml.Name{Local: "TEI"},
		Xmlns:   "http://www.tei-c.org/ns/1.0",
		Header: model.Header{
			FileDesc: model.FileDesc{
				TitleStmt:       model.TitleStmt{Title: "Converted from ALTO"},
				PublicationStmt: model.PublicationStmt{P: "Unpublished research data"},
				SourceDesc:      model.SourceDesc{P: "Derived from ALTO OCR"},
			},
		},
		Facsimile: model.Facsimile{Surfaces: []model.Surface{}},
		Text:      model.Text{Body: model.Body{}},
	}

	if len(entities) > 0 {
		if pd := buildProfileDesc(entities); pd != nil {
			doc.Header.ProfileDesc = pd
		}
	}

	for _, pg := range a.Layout.Page {
		pageID := pg.ID
		sPage := sanitizeID(pageID)
		surfaceID := surfaceID(pageID)

		surf := model.Surface{
			XmlID: surfaceID,
			Facs:  imageUrl,
			N:     pageID,
			Zones: []model.Zone{},
		}

		doc.Facsimile.Surfaces = append(doc.Facsimile.Surfaces, surf)
		doc.Text.Body.PBs = append(doc.Text.Body.PBs, model.PB{
			Facs: "#" + surfaceID,
			N:    pageID,
		})

		// transcription div for this page
		transDiv := model.Div{
			Type: "transcription",
			N:    pageID,
			Abs:  []model.AB{},
		}

		// We keep block structure: one <ab> per ALTO TextBlock
		for _, blk := range pg.PrintSpace.TextBlocks {
			ab := model.AB{
				XmlID: blockID(blk.ID),
				Type:  "ocr-block",
				Segs:  []model.Seg{},
			}

			for i, ln := range blk.Lines {
				lineID := fmt.Sprintf("l%04d", i+1)
				segID := fmt.Sprintf("ln_%s_%s_%04d", sPage, sanitizeID(blk.ID), i+1)
				lineText := joinLineStrings(ln.Strings)
				occs := occByKey[occKey(pageID, entityBlockID, lineID)]

				ab.Segs = append(ab.Segs, model.Seg{
					XmlID:   segID,
					Content: buildInlineForLineModel(lineText, occs, nil),
				})
			}

			transDiv.Abs = append(transDiv.Abs, ab)
		}

		doc.Text.Body.Divs = append(doc.Text.Body.Divs, transDiv)
	}

	buildStandOff(doc, entities)
	return doc, nil
}

func joinLineStrings(ss []alto.AltoString) string {
	parts := make([]string, 0, len(ss))
	for _, s := range ss {
		if s.Content == "" {
			continue
		}
		parts = append(parts, s.Content)
	}
	return strings.Join(parts, " ")
}
