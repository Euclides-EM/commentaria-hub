package tei

import (
	"encoding/xml"
	"fmt"
	"sort"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov/tei/model"
	"github.com/samber/lo"
)

func BuildTEIFromLines(
	lines LinesInput,
	entities *EntitiesInput,
	imageUrls map[string]string,
) (*model.TEI, error) {
	// Stable ordering of keys
	keys := make([]string, 0, len(lines.LinesByKeys))
	for k := range lines.LinesByKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	occByKey := buildEntitiesOccurrences(entities)

	doc := &model.TEI{
		XMLName: xml.Name{Local: "TEI"},
		Xmlns:   "http://www.tei-c.org/ns/1.0",
		Header: model.Header{
			FileDesc: model.FileDesc{
				TitleStmt: model.TitleStmt{Title: "Converted from lines"},
				PublicationStmt: model.PublicationStmt{
					P: "Unpublished research data",
				},
				SourceDesc: model.SourceDesc{
					P: "Derived from extracted text lines",
				},
			},
		},
		Facsimile: model.Facsimile{Surfaces: []model.Surface{}},
		Text:      model.Text{Body: model.Body{}},
	}

	// Generic profiles (optional)
	if entities != nil && len(entities.Profiles) > 0 {
		doc.Header.ProfileDesc = buildProfileDesc(entities)
	}

	// Collect language keys in stable order
	langSet := make(map[string]bool)
	for _, key := range keys {
		for lang := range lines.LinesByKeys[key].Translations {
			langSet[lang] = true
		}
	}
	langs := lo.Keys(langSet)
	sort.Strings(langs)

	for _, pageKey := range keys {
		pageLines := lines.LinesByKeys[pageKey]

		sKey := sanitizeID(pageKey)
		surfaceID := "page_" + sKey

		img := ""
		if imageUrls != nil {
			img = imageUrls[pageKey]
		}

		// facsimile surface (no zones here because we only have lines, not bbox)
		doc.Facsimile.Surfaces = append(doc.Facsimile.Surfaces, model.Surface{
			XmlID: surfaceID,
			Facs:  img,
			N:     pageKey,
			Zones: []model.Zone{},
		})

		// page break for this key
		doc.Text.Body.PBs = append(doc.Text.Body.PBs, model.PB{
			Facs: "#" + surfaceID,
			N:    pageKey,
		})

		// --- transcription div ---
		const blockID = "b1"

		transDiv := model.Div{
			Type:    "transcription",
			XmlLang: "", // set if you want, eg "fr"
			N:       pageKey,
			Abs: []model.AB{
				{
					XmlID: "b1_" + sKey,
					Type:  "transcription-block",
					Segs:  []model.Seg{},
				},
			},
		}

		for i, line := range pageLines.TranscriptionLines {
			lineID := fmt.Sprintf("l%04d", i+1)
			segID := segXMLID(pageKey, i+1)

			occs := occByKey[occKey(pageKey, blockID, lineID)]
			content := buildInlineForLineModel(line, occs)

			transDiv.Abs[0].Segs = append(transDiv.Abs[0].Segs, model.Seg{
				XmlID:   segID,
				Content: content,
			})
		}

		doc.Text.Body.Divs = append(doc.Text.Body.Divs, transDiv)
		// --- translation divs ---
		for _, lang := range langs {
			trLines := pageLines.Translations[lang]
			if len(trLines) == 0 {
				continue
			}

			// Enforce 1:1 mapping. If mismatch, fail loudly.
			if len(trLines) != len(pageLines.TranscriptionLines) {
				return nil, fmt.Errorf("translation line mismatch: lang=%s key=%s have=%d want=%d",
					lang, pageKey, len(trLines), len(pageLines.TranscriptionLines))
			}

			trDiv := model.Div{
				Type:    "translation",
				XmlLang: lang,
				N:       pageKey,
				Abs: []model.AB{
					{
						XmlID: "tr_" + sanitizeID(lang) + "_" + sKey,
						Type:  "translation-block",
						Segs:  []model.Seg{},
					},
				},
			}

			for i, tline := range trLines {
				corresp := "#" + segXMLID(pageKey, i+1)
				trDiv.Abs[0].Segs = append(trDiv.Abs[0].Segs, model.Seg{
					Corresp: corresp,
					Content: []model.Inline{{Text: tline}},
				})
			}

			doc.Text.Body.Divs = append(doc.Text.Body.Divs, trDiv)
		}

	}

	return doc, nil
}

func segXMLID(key string, oneBasedIdx int) string {
	return fmt.Sprintf("ln_%s_%04d", sanitizeID(key), oneBasedIdx)
}

func buildInlineForLineModel(line string, occs []EntityOccurrence) []model.Inline {
	if line == "" {
		return nil
	}
	if len(occs) == 0 {
		return []model.Inline{{Text: line}}
	}

	bLen := len(line)
	valid := make([]EntityOccurrence, 0, len(occs))
	for _, oc := range occs {
		if oc.Start < 0 || oc.End <= oc.Start {
			continue
		}
		if oc.Start >= bLen {
			continue
		}
		if oc.End > bLen {
			oc.End = bLen
		}
		valid = append(valid, oc)
	}
	if len(valid) == 0 {
		return []model.Inline{{Text: line}}
	}

	// remove overlaps (simple, non-nesting)
	outOccs := make([]EntityOccurrence, 0, len(valid))
	curEnd := -1
	for _, oc := range valid {
		if oc.Start < curEnd {
			continue
		}
		outOccs = append(outOccs, oc)
		curEnd = oc.End
	}

	out := make([]model.Inline, 0, 2*len(outOccs)+1)
	cursor := 0
	for _, oc := range outOccs {
		if cursor < oc.Start {
			out = append(out, model.Inline{Text: line[cursor:oc.Start]})
		}

		entityText := safeSliceByByteRange(line, oc.Start, oc.End)
		out = append(out, model.Inline{
			XMLName: xml.Name{Local: oc.Element},
			Text:    entityText,
			Ref:     oc.Ref,
			Ana:     oc.Ana,
		})

		cursor = oc.End
	}
	if cursor < len(line) {
		out = append(out, model.Inline{Text: line[cursor:]})
	}

	return mergeAdjacentTextModel(out)
}

func mergeAdjacentTextModel(in []model.Inline) []model.Inline {
	if len(in) < 2 {
		return in
	}
	out := make([]model.Inline, 0, len(in))
	for _, el := range in {
		if el.XMLName.Local == "" && el.Text != "" {
			if len(out) > 0 && out[len(out)-1].XMLName.Local == "" {
				out[len(out)-1].Text += el.Text
				continue
			}
		}
		out = append(out, el)
	}
	return out
}
