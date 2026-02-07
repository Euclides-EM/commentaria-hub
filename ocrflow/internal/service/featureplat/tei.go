package featureplat

import (
	"fmt"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store/featureplat"
)

type TEI struct {
	resultSvc              *Result
	tpsTranscriptionsStore *fpstore.TPSTranscriptions
}

func NewTEI(resultSvc *Result, tpsTranscriptionsStore *fpstore.TPSTranscriptions) *TEI {
	return &TEI{
		resultSvc:              resultSvc,
		tpsTranscriptionsStore: tpsTranscriptionsStore,
	}
}

func (t *TEI) GetTEI(collectionId, key string, featureParams []string) ([]byte, error) {

	// todo: current impl handles overlaps badly + no normalized + multiple occurrences not really working...

	if collectionId != "tps" {
		return nil, fmt.Errorf("TEI generation only supported for collection 'tps'")
	}

	if key == "" {
		return nil, fmt.Errorf("key parameter is required")
	}

	// Read CSV to get title and imprint
	title, imprint, err := t.tpsTranscriptionsStore.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcription for key %s: %w", key, err)
	}

	// Parse features filter (supports repeated `features` params and comma-separated lists)
	var featureFilter []string
	for _, raw := range featureParams {
		if raw == "" {
			continue
		}
		parts := strings.Split(raw, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				featureFilter = append(featureFilter, trimmed)
			}
		}
	}

	// Get feature results for this key
	results, err := t.resultSvc.ListResults(collectionId, []string{key}, featureFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature results: %w", err)
	}

	// Generate TEI XML
	teiXML := t.generateTEI(key, title, imprint, results)
	return []byte(teiXML), nil
}

func (t *TEI) generateTEI(key, title, imprint string, results []*featureplat.FeatureResult) string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<TEI xmlns="http://www.tei-c.org/ns/1.0">` + "\n\n")

	// teiHeader
	sb.WriteString("    <teiHeader>\n")
	sb.WriteString("        <fileDesc>\n")
	sb.WriteString("            <titleStmt>\n")
	sb.WriteString(fmt.Sprintf("                <title>Annotated title page: %s</title>\n", key))

	// Collect unique respStmt entries from results
	respStmts := t.collectRespStmts(results)
	for i, respStmt := range respStmts {
		sb.WriteString(fmt.Sprintf("                <respStmt xml:id=\"resp-%d\">\n", i+1))
		sb.WriteString(fmt.Sprintf("                    <resp>%s</resp>\n", respStmt.Resp))
		if respStmt.Name != "" {
			sb.WriteString(fmt.Sprintf("                    <name>%s</name>\n", xmlEscape(respStmt.Name)))
		}
		if respStmt.Feature != "" {
			sb.WriteString(fmt.Sprintf("                    <idno type=\"feature\">%s</idno>\n", xmlEscape(respStmt.Feature)))
		}
		if respStmt.Revision != "" {
			sb.WriteString(fmt.Sprintf("                    <idno type=\"revision\">%s</idno>\n", xmlEscape(respStmt.Revision)))
		}
		if respStmt.Id != "" {
			sb.WriteString(fmt.Sprintf("                    <idno type=\"id\">%s</idno>\n", xmlEscape(respStmt.Id)))
		}
		sb.WriteString("                </respStmt>\n")
	}

	sb.WriteString("            </titleStmt>\n")
	sb.WriteString("        </fileDesc>\n")
	sb.WriteString("    </teiHeader>\n\n")

	// text/body - build text with anchors for annotations
	sb.WriteString("    <text>\n")
	sb.WriteString("        <body>\n")
	sb.WriteString("            <p xml:id=\"p1\">\n")

	// Combine title and imprint
	combinedText := title
	if imprint != "" {
		combinedText += "\n" + imprint
	}

	// Build annotated text with anchors
	annotatedText, anchors := t.buildAnnotatedText(combinedText, results)
	sb.WriteString("                " + annotatedText)
	sb.WriteString("\n            </p>\n")
	sb.WriteString("        </body>\n")
	sb.WriteString("    </text>\n\n")

	// standOff with spans
	sb.WriteString("    <standOff>\n")
	sb.WriteString("        <spanGrp type=\"highlight\">\n")

	// Generate spans from feature results using anchors
	spanID := 1
	for anchorID, anchorInfo := range anchors {
		spanXML := t.generateSpanFromAnchor(spanID, anchorID, anchorInfo, respStmts)
		if spanXML != "" {
			sb.WriteString(spanXML)
			spanID++
		}
	}

	sb.WriteString("        </spanGrp>\n")
	sb.WriteString("    </standOff>\n\n")
	sb.WriteString("</TEI>\n")

	return sb.String()
}

type respStmtInfo struct {
	Resp     string
	Name     string
	Feature  string
	Revision string
	Id       string
}

func (t *TEI) collectRespStmts(results []*featureplat.FeatureResult) []respStmtInfo {
	seen := make(map[string]bool)
	var respStmts []respStmtInfo

	for _, result := range results {
		key := fmt.Sprintf("%s:%s:%s:%s:%s", result.Source.Resp, result.Source.Name, result.Source.Id, result.Source.Revision, result.Feature)
		if !seen[key] {
			seen[key] = true
			respStmts = append(respStmts, respStmtInfo{
				Resp:     result.Source.Resp,
				Name:     result.Source.Name,
				Feature:  result.Feature,
				Revision: result.Source.Revision,
				Id:       result.Source.Id,
			})
		}
	}

	return respStmts
}

type anchorInfo struct {
	result   *featureplat.FeatureResult
	value    featureplat.FeatureResultValue
	text     string
	startPos int
	endPos   int
}

func (t *TEI) buildAnnotatedText(text string, results []*featureplat.FeatureResult) (string, map[string]anchorInfo) {
	anchors := make(map[string]anchorInfo)
	anchorCounter := 1

	// Collect all text values from results
	type textMatch struct {
		text     string
		result   *featureplat.FeatureResult
		value    featureplat.FeatureResultValue
		startPos int
		endPos   int
	}
	var matches []textMatch

	for _, result := range results {
		for _, value := range result.Values {
			// Extract text from value (could be Root or from Childrens)
			valueText := value.Root
			if valueText == "" && len(value.Childrens) > 0 {
				// Try to get text from first child
				valueText = value.Childrens[0].Root
			}
			if valueText == "" {
				continue
			}

			// Try to find this text in the combined text (case-insensitive, but preserve original case)
			textLower := strings.ToLower(text)
			valueLower := strings.ToLower(valueText)
			pos := strings.Index(textLower, valueLower)
			if pos >= 0 {
				// Use the actual text from source (not normalized) for matching length
				actualLength := len(valueText)
				matches = append(matches, textMatch{
					text:     valueText,
					result:   result,
					value:    value,
					startPos: pos,
					endPos:   pos + actualLength,
				})
			}
		}
	}

	// Sort matches by start position
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].startPos > matches[j].startPos {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Build annotated text by inserting anchors (handle overlaps by taking first match)
	var parts []string
	lastPos := 0

	for _, match := range matches {
		// Skip if this match overlaps with previous (start before previous end)
		if match.startPos < lastPos {
			continue
		}

		// Add text before match
		if match.startPos > lastPos {
			parts = append(parts, xmlEscape(text[lastPos:match.startPos]))
		}

		// Add anchor start
		anchorID := fmt.Sprintf("anchor-%d", anchorCounter)
		anchorStartID := anchorID + "-s"
		anchorEndID := anchorID + "-e"
		parts = append(parts, fmt.Sprintf("<anchor xml:id=\"%s\"/>", anchorStartID))

		// Add matched text (use original text from source, not normalized)
		matchedText := text[match.startPos:match.endPos]
		parts = append(parts, xmlEscape(matchedText))

		// Add anchor end
		parts = append(parts, fmt.Sprintf("<anchor xml:id=\"%s\"/>", anchorEndID))

		// Store anchor info
		anchors[anchorID] = anchorInfo{
			result:   match.result,
			value:    match.value,
			text:     match.text,
			startPos: match.startPos,
			endPos:   match.endPos,
		}

		anchorCounter++
		lastPos = match.endPos
	}

	// Add remaining text
	if lastPos < len(text) {
		parts = append(parts, xmlEscape(text[lastPos:]))
	}

	annotatedText := strings.Join(parts, "")
	return annotatedText, anchors
}

func (t *TEI) generateSpanFromAnchor(spanID int, anchorID string, info anchorInfo, respStmts []respStmtInfo) string {
	// Find matching respStmt
	respID := ""
	for i, rs := range respStmts {
		if rs.Resp == info.result.Source.Resp &&
			rs.Name == info.result.Source.Name &&
			rs.Id == info.result.Source.Id &&
			rs.Revision == info.result.Source.Revision &&
			rs.Feature == info.result.Feature {
			respID = fmt.Sprintf("resp-%d", i+1)
			break
		}
	}
	if respID == "" {
		respID = "resp-1" // fallback
	}

	normalized := info.value.Root
	if normalized == "" && len(info.value.Childrens) > 0 {
		normalized = info.value.Childrens[0].Root
	}
	if normalized == "" {
		normalized = info.text
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("            <span xml:id=\"h-%s-%d\"\n", info.result.Feature, spanID))
	sb.WriteString(fmt.Sprintf("                  from=\"#%s-s\"\n", anchorID))
	sb.WriteString(fmt.Sprintf("                  to=\"#%s-e\"\n", anchorID))
	sb.WriteString(fmt.Sprintf("                  resp=\"#%s\"\n", respID))
	sb.WriteString("                  cert=\"medium\">\n")
	sb.WriteString("                <fs>\n")
	sb.WriteString(fmt.Sprintf("                    <f name=\"normalized\"><string>%s</string></f>\n", xmlEscape(normalized)))
	sb.WriteString("                </fs>\n")
	sb.WriteString("            </span>\n\n")

	return sb.String()
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
