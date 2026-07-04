package tei

import (
	"strconv"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

func ExtractTranslationLines(tei *model.TEI) []string {
	return extractDivLinesByType(tei, translationDivType)
}

func ExtractTranscriptionLines(tei *model.TEI) []string {
	return extractDivLinesByType(tei, transcriptionDivType)
}

func ExtractBiblMetadataLines(t *model.TEI) []string {
	if t == nil {
		return nil
	}

	var vals []string

	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		vals = append(vals, s)
	}

	h := t.Header
	add(h.FileDesc.TitleStmt.Title)
	add(h.FileDesc.PublicationStmt.Publisher)
	add(h.FileDesc.PublicationStmt.P)

	sd := &h.FileDesc.SourceDesc
	add(sd.P)

	if sd.BiblFull != nil {
		bf := sd.BiblFull

		if bf.TitleStmt != nil {
			for _, t := range bf.TitleStmt.Titles {
				add(t.Content)
			}
			for _, e := range bf.TitleStmt.Editor {
				add(e)
			}
		}

		if bf.PublicationStmt != nil {
			add(bf.PublicationStmt.PubPlace)
			add(bf.PublicationStmt.Publisher)

			if bf.PublicationStmt.Date != nil {
				add(bf.PublicationStmt.Date.When)
				add(bf.PublicationStmt.Date.Text)
			}

			for _, p := range bf.PublicationStmt.Ps {
				add(p.Text)
			}
		}

		if bf.Extent != nil {
			for _, m := range bf.Extent.Measures {
				if m.Quantity != 0 {
					add(strconv.Itoa(m.Quantity))
				}
				add(m.Text)
			}
		}

		if bf.NotesStmt != nil {
			for _, n := range bf.NotesStmt.Notes {
				add(n.Text)
			}
		}
	}

	return vals
}

func extractDivLinesByType(tei *model.TEI, divType string) []string {
	var lines []string
	for _, div := range tei.Text.Body.Divs {
		if div.Type == divType {
			for _, ab := range div.Abs {
				for _, l := range ab.Lines {
					lineText := extractTextFromNodes(l.Nodes)
					lines = append(lines, lineText)
				}
			}
		}
	}

	return lines
}

func extractTextFromNodes(nodes []model.ABNode) string {
	var text string
	for _, node := range nodes {
		if node.Inline != nil {
			text += node.Inline.Text
		} else {
			text += node.CharData
		}
	}
	return text
}
