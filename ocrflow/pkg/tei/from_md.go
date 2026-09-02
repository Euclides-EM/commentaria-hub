package tei

import (
	"encoding/xml"
	"regexp"
	"strconv"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/markdown"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

var (
	markdownZoneStartPattern  = regexp.MustCompile(`^\[(Margin|Footnote|Handwritten|Other)(?: type="([^"]+)")?\]$`)
	markdownDropcapPattern    = regexp.MustCompile(`^\{dropcap:([^|}]+)\|lines=([^|}]+)\|style=(plain|decorated|unknown)(?:\|decoration="([^"]*)")?\}`)
	markdownCorrectionPattern = regexp.MustCompile(`^\{printer-error-correction:([^}]+)\}`)
	markdownIllegiblePattern  = regexp.MustCompile(`^\[illegible(?:: ([^\]]+))?\]`)
	markdownUnclearPattern    = regexp.MustCompile(`^\[unclear: ([^\]]+)\]`)
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

		if blockType, content, ok := parseMarkdownFurniture(trimmed); ok {
			flushParagraph()
			abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, blockType, []string{content}))
			continue
		}

		if level, content := markdown.ParseHeader(line); level > 0 {
			flushParagraph()
			abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, markdown.HeaderPrefix+strconv.Itoa(level), []string{content}))
			continue
		}

		if isMarkdownTableStart(lines, i) {
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

		if blockType, closingTag, ok := parseMarkdownContainerStart(trimmed); ok {
			flushParagraph()
			var content []string
			for i++; i < len(lines) && strings.TrimSpace(lines[i]) != closingTag; i++ {
				content = append(content, lines[i])
			}
			abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, blockType, content))
			continue
		}

		if blockType, description, ok := parseMarkdownPageObject(trimmed); ok {
			flushParagraph()
			abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, blockType, []string{description}))
			continue
		}

		if trimmed == "[Blank page]" {
			flushParagraph()
			abs = append(abs, newMarkdownAB(pageKey, len(abs)+1, "blank-page", []string{""}))
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
		dropcapStart := strings.Index(s, "{dropcap:")
		correctionStart := strings.Index(s, "{printer-error-correction:")
		next := firstMarkdownInlineIndex(linkStart, boldStart, italicStart, dropcapStart, correctionStart)
		if next < 0 {
			nodes = appendTextNode(nodes, s)
			break
		}
		nodes = appendTextNode(nodes, s[:next])
		s = s[next:]

		switch {
		case strings.HasPrefix(s, "{dropcap:"):
			match := markdownDropcapPattern.FindStringSubmatch(s)
			if match == nil {
				nodes = appendTextNode(nodes, s[:1])
				s = s[1:]
				continue
			}
			rend := "dropcap lines=" + match[2] + " style=" + match[3]
			if match[4] != "" {
				rend += " decoration=" + match[4]
			}
			nodes = append(nodes, model.ABNode{Inline: &model.Inline{
				XMLName: xml.Name{Local: "hi"}, Rend: rend, Text: match[1],
			}})
			s = s[len(match[0]):]
		case strings.HasPrefix(s, "{printer-error-correction:"):
			match := markdownCorrectionPattern.FindStringSubmatch(s)
			if match == nil {
				nodes = appendTextNode(nodes, s[:1])
				s = s[1:]
				continue
			}
			nodes = append(nodes, model.ABNode{Inline: &model.Inline{
				XMLName: xml.Name{Local: "note"}, Ana: "printer-error-correction", Text: " [correction: " + match[1] + "]",
			}})
			s = s[len(match[0]):]
		case strings.HasPrefix(s, "[illegible"):
			match := markdownIllegiblePattern.FindStringSubmatch(s)
			if match == nil {
				nodes = appendTextNode(nodes, s[:1])
				s = s[1:]
				continue
			}
			nodes = append(nodes, model.ABNode{Inline: &model.Inline{
				XMLName: xml.Name{Local: "gap"}, Ana: "illegible", Text: match[0],
			}})
			s = s[len(match[0]):]
		case strings.HasPrefix(s, "[unclear:"):
			match := markdownUnclearPattern.FindStringSubmatch(s)
			if match == nil {
				nodes = appendTextNode(nodes, s[:1])
				s = s[1:]
				continue
			}
			nodes = append(nodes, model.ABNode{Inline: &model.Inline{
				XMLName: xml.Name{Local: "unclear"}, Text: match[0],
			}})
			s = s[len(match[0]):]
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

func parseMarkdownFurniture(line string) (string, string, bool) {
	for label, blockType := range map[string]string{
		"Running title": "running-title",
		"Page number":   "page-number",
		"Signature":     "signature",
		"Catchword":     "catchword",
	} {
		prefix := "<!-- " + label + ": "
		if strings.HasPrefix(line, prefix) && strings.HasSuffix(line, " -->") {
			return blockType, strings.TrimSuffix(strings.TrimPrefix(line, prefix), " -->"), true
		}
	}
	return "", "", false
}

func parseMarkdownContainerStart(line string) (string, string, bool) {
	if line == "[Calculation]" {
		return "calculation", "[/Calculation]", true
	}
	match := markdownZoneStartPattern.FindStringSubmatch(line)
	if match == nil {
		return "", "", false
	}
	blockType := strings.ToLower(match[1])
	if match[2] != "" {
		blockType += ":" + match[2]
	}
	return blockType, "[/" + match[1] + "]", true
}

func parseMarkdownPageObject(line string) (string, string, bool) {
	for _, label := range []string{"Diagram", "Figure", "Illustration", "Ornament"} {
		if line == "["+label+"]" {
			return strings.ToLower(label), "", true
		}
		prefix := "[" + label + ": "
		if strings.HasPrefix(line, prefix) && strings.HasSuffix(line, "]") {
			return strings.ToLower(label), strings.TrimSuffix(strings.TrimPrefix(line, prefix), "]"), true
		}
	}
	return "", "", false
}

func isMarkdownTableStart(lines []string, index int) bool {
	if index+1 >= len(lines) || !isMarkdownTableLine(strings.TrimSpace(lines[index])) {
		return false
	}
	return isMarkdownTableSeparator(strings.TrimSpace(lines[index+1]))
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
	var cells []string
	var cell strings.Builder
	escaped := false
	for _, r := range line {
		if escaped {
			if r != '|' {
				cell.WriteRune('\\')
			}
			cell.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '|' {
			cells = append(cells, cell.String())
			cell.Reset()
			continue
		}
		cell.WriteRune(r)
	}
	if escaped {
		cell.WriteRune('\\')
	}
	cells = append(cells, cell.String())
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return strings.Join(cells, " | ")
}
