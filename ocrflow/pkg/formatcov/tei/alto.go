package tei

import (
	"encoding/xml"
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov/tei/model"
)

func BuildTEIFromALTO(
	a *alto.Alto,
	entities *EntitiesInput,
	imageUrl string,
) (*model.TEI, error) {

	if a == nil {
		return nil, fmt.Errorf("alto is nil")
	}

	occByKey := buildEntitiesOccurrences(entities)

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

	if entities != nil && len(entities.Profiles) > 0 {
		doc.Header.ProfileDesc = buildProfileDesc(entities)
	}

	for _, pg := range a.Layout.Page {
		pageID := pg.ID
		sPage := sanitizeID(pageID)
		surfaceID := "page_" + sPage

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
				XmlID: "b_" + sanitizeID(blk.ID),
				Type:  "ocr-block",
				Segs:  []model.Seg{},
			}

			const blockID = "b1" // for entity targeting in this builder, we use a fixed block id
			for i, ln := range blk.Lines {
				lineID := fmt.Sprintf("l%04d", i+1)
				segID := fmt.Sprintf("ln_%s_%s_%04d", sPage, sanitizeID(blk.ID), i+1)

				lineText := joinLineStrings(ln.Strings)
				occs := occByKey[occKey(pageID, blockID, lineID)]

				ab.Segs = append(ab.Segs, model.Seg{
					XmlID:   segID,
					Content: buildInlineForLineModel(lineText, occs),
				})
			}

			transDiv.Abs = append(transDiv.Abs, ab)
		}

		doc.Text.Body.Divs = append(doc.Text.Body.Divs, transDiv)
	}

	return doc, nil
}
