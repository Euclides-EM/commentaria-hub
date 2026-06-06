package tei

import (
	"encoding/xml"
	"sort"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
	"github.com/samber/lo"
)

// BuildTEIFromLines builds TEI for a single page. pageKey identifies the page (e.g. "1", "page1").
func BuildTEIFromLines(
	pageKey string,
	lines Lines,
	entities []EntityItem,
	imageUrl string,
	biblMetadata *model.BiblFull,
) (*model.TEI, error) {
	doc := &model.TEI{
		XMLName: xml.Name{Local: "TEI"},
		Xmlns:   "http://www.tei-c.org/ns/1.0",
		Header: model.Header{
			FileDesc: buildFileDesc(biblMetadata),
			StandOff: buildStandOff(pageKey, entities, nil, alto.Tags{}),
		},
		Facsimile: buildFacsimileForLines(pageKey, imageUrl, lines.TranscriptionLines),
		Text:      model.Text{Body: model.Body{}},
	}

	doc.Text.Body.PBs = append(doc.Text.Body.PBs, model.PB{
		Facs: "#" + doc.Facsimile.Surfaces[0].XmlID,
		N:    pageKey,
	})

	ab := model.AB{
		XmlID: transcriptionAnonBlockID(pageKey, 1),
		Type:  transcriptionAnonBlockType,
		Facs:  "#" + facZoneBlockID(pageKey, 1),
	}

	for i, line := range lines.TranscriptionLines {
		lineKey := strconv.Itoa(i)
		nodes := buildInlineNodesWithAnchors("1", lineKey, line, entities)
		l := model.L{
			XmlID: lineID(pageKey, 1, i+1),
			Facs:  "#" + facZoneLineID(pageKey, 1, i+1),
			Nodes: nodes,
		}
		ab.Lines = append(ab.Lines, l)
	}

	transDiv := model.Div{
		Type: transcriptionDivType,
		N:    pageKey,
		Abs:  []model.AB{ab},
	}
	doc.Text.Body.Divs = append(doc.Text.Body.Divs, transDiv)

	// Collect language keys in stable order
	langSet := make(map[string]bool)
	for lang := range lines.Translations {
		langSet[lang] = true
	}
	langs := lo.Keys(langSet)
	sort.Strings(langs)
	// --- translation divs ---
	for _, lang := range langs {
		trLines := lines.Translations[lang]
		if len(trLines) == 0 {
			continue
		}

		addCorresp := len(trLines) == len(lines.TranscriptionLines)
		trDiv := model.Div{
			Type:    translationDivType,
			XmlLang: lang,
			N:       pageKey,
			Abs: []model.AB{
				{
					XmlID: translationAnonBlockID(pageKey, lang, 1),
					Type:  translationAnonBlockType,
					Lines: lo.Map(lines.Translations[lang], func(line string, i int) model.L {
						l := model.L{
							Nodes: buildInlineNodesWithAnchors("1", strconv.Itoa(i), line, nil),
						}
						if addCorresp {
							l.Corresp = "#" + lineID(pageKey, 1, i+1)
						}
						return l
					}),
				},
			},
		}

		doc.Text.Body.Divs = append(doc.Text.Body.Divs, trDiv)
	}

	return doc, nil
}
