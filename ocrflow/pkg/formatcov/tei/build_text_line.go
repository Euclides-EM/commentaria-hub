package tei

import (
	"encoding/xml"
	"strings"
	"unicode/utf8"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov/tei/model"
)

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

func buildInlineForLine(line string, occs []EntityOccurrence) []model.Inline {
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

	return mergeAdjacentText(out)
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

func mergeAdjacentText(in []model.Inline) []model.Inline {
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
