package tei

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
	"github.com/samber/lo"
)

// BuildTEIFromLines builds TEI for a single page. pageKey identifies the page (e.g. "1", "page1").
func BuildTEIFromLines(
	pageKey string,
	lines Lines,
	entities []EntityItem,
	imageUrl string,
) (*model.TEI, error) {
	pageLines := lines.TranscriptionLines
	orderedKeys := make([]string, 0, len(pageLines))
	lineTextByKey := make(map[string]string)
	for i := 1; i <= len(pageLines); i++ {
		k := occKey(pageKey, "b1", fmt.Sprintf("l%04d", i))
		orderedKeys = append(orderedKeys, k)
		lineTextByKey[k] = pageLines[i-1]
	}
	getLineText := func(key string) string { return lineTextByKey[key] }
	occByKey := buildEntitiesOccurrences(entities, orderedKeys, getLineText)

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

	if entities != nil && len(entities) > 0 {
		if ed := buildEncodingDesc(entities); ed != nil {
			doc.Header.EncodingDesc = ed
		}
		if pd := buildProfileDesc(entities); pd != nil {
			doc.Header.ProfileDesc = pd
		}
	}

	// Stand-off: build mentions list from occurrences, then spanGrp + listRelation (feature-assignment facts).
	mentionsForStandOff := flattenMentionsForStandOff(orderedKeys, occByKey)
	if len(mentionsForStandOff) > 0 {
		buildStandOffMentions(doc, mentionsForStandOff)
	} else if entities != nil && len(entities) > 0 {
		buildStandOff(doc, entities)
	}

	// Collect language keys in stable order
	langSet := make(map[string]bool)
	for lang := range lines.Translations {
		langSet[lang] = true
	}
	langs := lo.Keys(langSet)
	sort.Strings(langs)

	sKey := sanitizeID(pageKey)
	surfaceID := surfaceID(pageKey)

	blockZoneID := fmt.Sprintf("z_blk_%s_1", sKey)
	zones := []model.Zone{
		{XmlID: blockZoneID, Type: "block", ULX: 0, ULY: 0, LRX: 0, LRY: 0},
	}
	for i := 1; i <= len(pageLines); i++ {
		segID := segXMLID(pageKey, i)
		lineZoneID := "z_" + segID
		zones = append(zones, model.Zone{
			XmlID:   lineZoneID,
			Type:    "line",
			Corresp: "#" + blockZoneID,
			ULX:     0, ULY: 0, LRX: 0, LRY: 0,
		})
	}

	doc.Facsimile.Surfaces = append(doc.Facsimile.Surfaces, model.Surface{
		XmlID: surfaceID,
		Facs:  imageUrl,
		N:     pageKey,
		Zones: zones,
	})

	doc.Text.Body.PBs = append(doc.Text.Body.PBs, model.PB{
		Facs: "#" + surfaceID,
		N:    pageKey,
	})

	const blockID = "b1"
	ab := model.AB{
		XmlID: "b1_" + sKey,
		Type:  "transcription-block",
		Facs:  "#" + blockZoneID,
		Nodes: []model.ABNode{},
	}

	for i, line := range pageLines {
		lineID := fmt.Sprintf("l%04d", i+1)
		segID := segXMLID(pageKey, i+1)
		lineZoneID := "z_" + segID
		occs := occByKey[occKey(pageKey, blockID, lineID)]
		nodes := buildInlineNodesWithAnchors(line, occs)
		for _, nd := range nodes {
			ab.Nodes = append(ab.Nodes, nd)
		}
		ab.Nodes = append(ab.Nodes, model.ABNode{
			LB: &model.LB{XmlID: segID, Facs: "#" + lineZoneID},
		})
	}

	transDiv := model.Div{
		Type: "transcription",
		N:    pageKey,
		Abs:  []model.AB{ab},
	}

	doc.Text.Body.Divs = append(doc.Text.Body.Divs, transDiv)
	// --- translation divs ---
	for _, lang := range langs {
		trLines := lines.Translations[lang]
		if len(trLines) == 0 {
			continue
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

		if len(trLines) == len(pageLines) {
			for i, tline := range trLines {
				corresp := "#" + segXMLID(pageKey, i+1)
				trDiv.Abs[0].Segs = append(trDiv.Abs[0].Segs, model.Seg{
					Corresp: corresp,
					Content: []model.Inline{{Text: tline}},
				})
			}
		} else {
			// If translation lines don't match transcription lines, just add them without corresp
			for _, tline := range trLines {
				trDiv.Abs[0].Segs = append(trDiv.Abs[0].Segs, model.Seg{
					Content: []model.Inline{{Text: tline}},
				})
			}
		}

		doc.Text.Body.Divs = append(doc.Text.Body.Divs, trDiv)
	}

	return doc, nil
}

// flattenMentionsForStandOff returns mentions in document order for buildStandOffMentions.
func flattenMentionsForStandOff(orderedKeys []string, occByKey map[string][]mentionInLine) []MentionForStandOff {
	var out []MentionForStandOff
	for _, k := range orderedKeys {
		for _, oc := range occByKey[k] {
			out = append(out, MentionForStandOff{
				MentionID: oc.MentionID,
				Ref:       oc.Ref,
				Ana:       oc.Ana,
			})
		}
	}
	return out
}

func segXMLID(key string, oneBasedIdx int) string {
	return fmt.Sprintf("ln_%s_%04d", sanitizeID(key), oneBasedIdx)
}

// anchorEvent is a start or end of a mention span at a byte position.
type anchorEvent struct {
	pos   int
	isEnd bool
	oc    mentionInLine
}

// buildInlineNodesWithAnchors builds ABNodes with anchor pairs around mention text (stand-off style).
// Overlapping mentions are supported: all get anchors; at each position we emit end anchors then start anchors, then text.
func buildInlineNodesWithAnchors(line string, occs []mentionInLine) []model.ABNode {
	if line == "" {
		return nil
	}
	if len(occs) == 0 {
		return []model.ABNode{{CharData: line}}
	}

	bLen := len(line)
	var events []anchorEvent
	for _, oc := range occs {
		if oc.Start < 0 || oc.End <= oc.Start || oc.Start >= bLen {
			continue
		}
		end := oc.End
		if end > bLen {
			end = bLen
		}
		events = append(events, anchorEvent{pos: oc.Start, isEnd: false, oc: mentionInLine{Start: oc.Start, End: end, Ref: oc.Ref, Ana: oc.Ana, MentionID: oc.MentionID}})
		events = append(events, anchorEvent{pos: end, isEnd: true, oc: mentionInLine{Start: oc.Start, End: end, Ref: oc.Ref, Ana: oc.Ana, MentionID: oc.MentionID}})
	}
	if len(events) == 0 {
		return []model.ABNode{{CharData: line}}
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].pos != events[j].pos {
			return events[i].pos < events[j].pos
		}
		if events[i].isEnd != events[j].isEnd {
			return events[i].isEnd // end before start at same position
		}
		// Same position, same type: for ends, close inner (started later) first so a_m3_e before a_m1_e
		if events[i].isEnd {
			return events[i].oc.Start > events[j].oc.Start
		}
		return events[i].oc.Start < events[j].oc.Start
	})

	// Distinct positions so we can output text from p to nextP after start anchors at p
	posSet := make(map[int]struct{})
	for _, e := range events {
		posSet[e.pos] = struct{}{}
	}
	positions := make([]int, 0, len(posSet))
	for pos := range posSet {
		positions = append(positions, pos)
	}
	sort.Ints(positions)

	var out []model.ABNode
	pos := 0
	runStart := 0 // after end anchors, next run can start here (for " Smith" when gap is single char)
	for idx, p := range positions {
		nextP := len(line)
		if idx+1 < len(positions) {
			nextP = positions[idx+1]
		}
		hasStart := false
		for _, e := range events {
			if e.pos == p && !e.isEnd {
				hasStart = true
				break
			}
		}
		// If we have start anchors at p and runStart < p with a tiny gap (≤1 char), skip outputting gap here; it goes in the run after starts.
		if !(hasStart && runStart < p && p-runStart <= 1) && pos < p {
			out = append(out, model.ABNode{CharData: safeSliceByByteRange(line, pos, p)})
			runStart = p // next run starts at p
		}
		for _, e := range events {
			if e.pos != p {
				continue
			}
			if e.isEnd {
				anchorPart := strings.ReplaceAll(e.oc.MentionID, "m_", "m")
				out = append(out, model.ABNode{Anchor: &model.Anchor{XmlID: "a_" + anchorPart + "_e"}})
				runStart = p
			}
		}
		for _, e := range events {
			if e.pos != p || e.isEnd {
				continue
			}
			anchorPart := strings.ReplaceAll(e.oc.MentionID, "m_", "m")
			out = append(out, model.ABNode{Anchor: &model.Anchor{XmlID: "a_" + anchorPart + "_s"}})
		}
		if hasStart && runStart < nextP {
			out = append(out, model.ABNode{CharData: safeSliceByByteRange(line, runStart, nextP)})
			runStart = nextP
		}
		pos = nextP
	}
	if runStart < len(line) {
		out = append(out, model.ABNode{CharData: safeSliceByByteRange(line, runStart, len(line))})
	}
	return mergeAdjacentCharDataNodes(out)
}

func mergeAdjacentCharDataNodes(in []model.ABNode) []model.ABNode {
	if len(in) < 2 {
		return in
	}
	out := make([]model.ABNode, 0, len(in))
	for _, n := range in {
		if n.CharData != "" {
			if len(out) > 0 && out[len(out)-1].CharData != "" {
				out[len(out)-1].CharData += n.CharData
				continue
			}
		}
		out = append(out, n)
	}
	return out
}

func buildInlineForLineModel(line string, occs []mentionInLine, refToFactIDs map[string][]string) []model.Inline {
	if line == "" {
		return nil
	}
	if len(occs) == 0 {
		return []model.Inline{{Text: line}}
	}

	bLen := len(line)
	valid := make([]mentionInLine, 0, len(occs))
	for _, oc := range occs {
		if oc.Start < 0 || oc.End <= oc.Start {
			continue
		}
		if oc.Start >= bLen {
			continue
		}
		end := oc.End
		if end > bLen {
			end = bLen
		}
		valid = append(valid, mentionInLine{Start: oc.Start, End: end, Ref: oc.Ref, Ana: oc.Ana, MentionID: oc.MentionID})
	}
	if len(valid) == 0 {
		return []model.Inline{{Text: line}}
	}

	// remove overlaps (simple, non-nesting)
	outOccs := make([]mentionInLine, 0, len(valid))
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
		inline := model.Inline{
			XMLName: xml.Name{Local: "name"},
			Text:    entityText,
			Ref:     oc.Ref,
			Ana:     oc.Ana,
			XmlID:   oc.MentionID,
		}
		if refToFactIDs != nil && oc.Ref != "" {
			normRef := strings.TrimPrefix(oc.Ref, "#")
			if ids := refToFactIDs[normRef]; len(ids) > 0 {
				inline.Corresp = "#" + ids[0]
			}
		}
		out = append(out, inline)
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

func safeSliceByByteRange(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start >= end {
		return ""
	}
	// Keep UTF-8 safe boundaries
	for start > 0 && !utf8.ValidString(s[start:]) {
		start--
	}
	for end < len(s) && !utf8.ValidString(s[:end]) {
		end++
	}
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}
