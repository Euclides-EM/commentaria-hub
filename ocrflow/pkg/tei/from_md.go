package tei

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/markdown"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

func BuildTEIFromMarkdown(
	pageKey string,
	a *markdown.Markdown,
	biblMetadata *model.BiblFull,
) (*model.TEI, error) {
	doc := &model.TEI{
		XMLName: xml.Name{Local: "TEI"},
		Xmlns:   "http://www.tei-c.org/ns/1.0",
		Header: model.Header{
			FileDesc: buildFileDesc(biblMetadata),
		},
		Text: model.Text{Body: model.Body{}},
	}

	doc.Text.Body.PBs = append(doc.Text.Body.PBs, model.PB{N: pageKey})
	doc.Text.Body.Divs = append(doc.Text.Body.Divs, model.Div{
		Type: transcriptionDivType,
		N:    pageKey,
		Abs:  markdownBlocksToABs(pageKey, a),
	})
	return doc, nil
}

func markdownBlocksToABs(pageKey string, md *markdown.Markdown) []model.AB {
	if md == nil {
		return nil
	}

	lines := strings.Split(md.Content, "\n")
	var abs []model.AB
	var paragraph []string

	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}
		abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, "paragraph", paragraph))
		paragraph = nil
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			flushParagraph()
			continue
		}

		if comment, ok := parseMarkdownComment(trimmed); ok {
			flushParagraph()
			abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, "comment", []string{comment}))
			continue
		}

		if level, content := markdown.ParseHeader(line); level > 0 {
			flushParagraph()
			abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, markdown.HeaderPrefix+strconv.Itoa(level), []string{content}))
			continue
		}

		if isMarkdownTableLine(trimmed) {
			flushParagraph()
			var tableLines []string
			for i < len(lines) && isMarkdownTableLine(strings.TrimSpace(lines[i])) {
				if !isMarkdownTableSeparator(strings.TrimSpace(lines[i])) {
					tableLines = append(tableLines, markdownTableLineToText(lines[i]))
				}
				i++
			}
			i--
			abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, "table", tableLines))
			continue
		}

		if figure := parseMarkdownFigure(trimmed); figure != "" {
			flushParagraph()
			abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, "figure", []string{figure}))
			continue
		}

		paragraph = append(paragraph, line)
	}
	flushParagraph()
	return abs
}

func newMarkdownAB(pageKey string, blockIdx int, blockType string, lines []string) model.AB {
	ab := model.AB{
		XmlID: transcriptionAnonBlockID(pageKey, blockIdx),
		Type:  blockType,
	}
	for i, line := range lines {
		ab.Lines = append(ab.Lines, model.L{
			XmlID: lineID(pageKey, blockIdx, i+1),
			Nodes: markdownInlineNodes(line),
		})
	}
	return ab
}

func markdownInlineNodes(s string) []model.ABNode {
	var nodes []model.ABNode
	for len(s) > 0 {
		linkStart := strings.Index(s, "[")
		boldStart := strings.Index(s, "**")
		italicStart := strings.Index(s, "*")
		next := firstMarkdownInlineIndex(linkStart, boldStart, italicStart)
		if next < 0 {
			nodes = appendTextNode(nodes, s)
			break
		}
		nodes = appendTextNode(nodes, s[:next])
		s = s[next:]

		switch {
		case strings.HasPrefix(s, "["):
			endText := strings.Index(s, "](")
			if endText < 0 {
				nodes = appendTextNode(nodes, s[:1])
				s = s[1:]
				continue
			}
			endTarget := strings.Index(s[endText+2:], ")")
			if endTarget < 0 {
				nodes = appendTextNode(nodes, s[:1])
				s = s[1:]
				continue
			}
			text := s[1:endText]
			target := s[endText+2 : endText+2+endTarget]
			nodes = append(nodes, model.ABNode{Inline: &model.Inline{
				XMLName: xml.Name{Local: "ref"},
				Target:  target,
				Text:    text,
			}})
			s = s[endText+2+endTarget+1:]
		case strings.HasPrefix(s, "**"):
			end := strings.Index(s[2:], "**")
			if end < 0 {
				nodes = appendTextNode(nodes, s[:2])
				s = s[2:]
				continue
			}
			nodes = append(nodes, model.ABNode{Inline: &model.Inline{
				XMLName: xml.Name{Local: "hi"},
				Rend:    "bold",
				Text:    s[2 : 2+end],
			}})
			s = s[2+end+2:]
		case strings.HasPrefix(s, "*"):
			end := strings.Index(s[1:], "*")
			if end < 0 {
				nodes = appendTextNode(nodes, s[:1])
				s = s[1:]
				continue
			}
			nodes = append(nodes, model.ABNode{Inline: &model.Inline{
				XMLName: xml.Name{Local: "hi"},
				Rend:    "italic",
				Text:    s[1 : 1+end],
			}})
			s = s[1+end+1:]
		}
	}
	return nodes
}

func appendTextNode(nodes []model.ABNode, text string) []model.ABNode {
	if text == "" {
		return nodes
	}
	return append(nodes, model.ABNode{CharData: text})
}

func firstMarkdownInlineIndex(indexes ...int) int {
	min := -1
	for _, idx := range indexes {
		if idx < 0 {
			continue
		}
		if min < 0 || idx < min {
			min = idx
		}
	}
	return min
}

func parseMarkdownComment(line string) (string, bool) {
	if !strings.HasPrefix(line, "<!--") || !strings.HasSuffix(line, "-->") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "<!--"), "-->")), true
}

func parseMarkdownFigure(line string) string {
	if !strings.HasPrefix(line, "*[") || !strings.HasSuffix(line, "]*") {
		return ""
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "*["), "]*"))
}

func isMarkdownTableLine(line string) bool {
	return strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") && strings.Count(line, "|") >= 2
}

func isMarkdownTableSeparator(line string) bool {
	line = strings.Trim(line, "| ")
	if line == "" {
		return false
	}
	for _, r := range line {
		if r != '-' && r != ':' && r != '|' && r != ' ' {
			return false
		}
	}
	return strings.Contains(line, "-")
}

func markdownTableLineToText(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	cells := strings.Split(line, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return strings.Join(cells, "\t")
}
