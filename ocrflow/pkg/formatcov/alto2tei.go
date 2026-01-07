package formatcov

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/xml"
)

type ALTOToTEIOptions struct {
	RowTolPx     float64 // lines within this VPOS distance are treated as same row
	ParaGapPx    float64 // effective vertical gap that starts a new paragraph
	KeepEmpty    bool
	Title        string
	FacsFromPage bool // if true, pb@facs uses Page.ID
}

// ConvertALTOToTEI converts an ALTO document to TEI XML (as bytes).
func ConvertALTOToTEI(a *alto.Alto, opts ALTOToTEIOptions) ([]byte, error) {
	if opts.RowTolPx <= 0 {
		opts.RowTolPx = 6
	}
	if opts.ParaGapPx <= 0 {
		opts.ParaGapPx = 28
	}
	if opts.Title == "" {
		opts.Title = "Converted from ALTO"
	}
	if !opts.FacsFromPage {
		// default true, but keep the option explicit
		opts.FacsFromPage = true
	}

	var buf bytes.Buffer
	w := io.Writer(&buf)

	// Header
	writeString(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	writeString(w, `<TEI xmlns="http://www.tei-c.org/ns/1.0">`+"\n")
	writeString(w, `  <teiHeader>`+"\n")
	writeString(w, `    <fileDesc>`+"\n")
	writeString(w, fmt.Sprintf(`      <titleStmt><title>%s</title></titleStmt>`+"\n", xmlEscapeText(opts.Title)))
	writeString(w, `      <publicationStmt><p>Publication Information</p></publicationStmt>`+"\n")
	writeString(w, `      <sourceDesc><p>Information about the source</p></sourceDesc>`+"\n")
	writeString(w, `    </fileDesc>`+"\n")
	writeString(w, `  </teiHeader>`+"\n")
	writeString(w, `  <text>`+"\n")
	writeString(w, `    <body>`+"\n")

	// Pages
	for _, p := range a.Layout.Page {
		facs := ""
		if opts.FacsFromPage {
			facs = p.ID
		}
		writeString(w, fmt.Sprintf(`      <pb facs="%s"/>`+"\n", xmlEscapeAttr(facs)))

		lines := flattenLines(&p)
		sortLines(lines, opts.RowTolPx)

		paras := groupIntoParas(lines, opts.ParaGapPx, opts.KeepEmpty)
		for _, para := range paras {
			writeString(w, `      <p>`)
			for _, ln := range para {
				lbID := xml.SanitizeXMLID(ln.LineID)

				if lbID != "" {
					writeString(w, fmt.Sprintf(`<lb xml:id="%s"/>`, xmlEscapeAttr(lbID)))
				} else {
					writeString(w, `<lb/>`)
				}

				writeString(w, xmlEscapeText(ln.Text))
			}
			writeString(w, `</p>`+"\n")
		}
	}

	// Close
	writeString(w, `    </body>`+"\n")
	writeString(w, `  </text>`+"\n")
	writeString(w, `</TEI>`+"\n")

	return buf.Bytes(), nil
}

func flattenLines(p *alto.Page) []alto.Line {
	out := make([]alto.Line, 0, 1024)

	for _, b := range p.PrintSpace.TextBlocks {
		for _, tl := range b.Lines {
			txt := joinStrings(tl.Strings)
			txt = normalizeSpaces(txt)

			out = append(out, alto.Line{
				BlockID:  b.ID,
				TagRefs:  b.TagRefs,
				HPOS:     tl.HPOS,
				VPOS:     tl.VPOS,
				Height:   tl.Height,
				Text:     txt,
				LineID:   tl.ID,
				BlockVP:  float64(b.VPOS),
				BlockHP:  float64(b.HPOS),
				BlockHgt: float64(b.Height),
			})
		}
	}
	return out
}

func sortLines(lines []alto.Line, rowTol float64) {
	sort.SliceStable(lines, func(i, j int) bool {
		a, b := lines[i], lines[j]
		dy := a.VPOS - b.VPOS

		// Same row: left to right
		if math.Abs(dy) <= rowTol {
			if a.HPOS != b.HPOS {
				return a.HPOS < b.HPOS
			}
			// Fallbacks for stability
			if a.BlockVP != b.BlockVP {
				return a.BlockVP < b.BlockVP
			}
			if a.BlockHP != b.BlockHP {
				return a.BlockHP < b.BlockHP
			}
			return a.LineID < b.LineID
		}

		// Different rows: top to bottom
		return a.VPOS < b.VPOS
	})
}

func groupIntoParas(lines []alto.Line, paraGap float64, keepEmpty bool) [][]alto.Line {
	var paras [][]alto.Line
	var cur []alto.Line

	flush := func() {
		if len(cur) == 0 {
			return
		}
		if !keepEmpty {
			allEmpty := true
			for _, ln := range cur {
				if strings.TrimSpace(ln.Text) != "" {
					allEmpty = false
					break
				}
			}
			if allEmpty {
				cur = nil
				return
			}
		}
		paras = append(paras, cur)
		cur = nil
	}

	var prev *alto.Line
	for i := range lines {
		ln := lines[i]
		if !keepEmpty && strings.TrimSpace(ln.Text) == "" {
			continue
		}

		if prev == nil {
			cur = append(cur, ln)
			prev = &ln
			continue
		}

		// Effective gap accounts for line height
		dy := ln.VPOS - prev.VPOS
		effectiveGap := dy - prev.Height

		blockChanged := ln.BlockID != prev.BlockID

		// Paragraph break heuristics:
		// - clear vertical gap
		// - or block change with a moderate gap
		if effectiveGap > paraGap || (blockChanged && effectiveGap > paraGap/2) {
			flush()
		}

		cur = append(cur, ln)
		prev = &ln
	}

	flush()
	return paras
}

func joinStrings(ss []alto.AltoString) string {
	if len(ss) == 0 {
		return ""
	}
	var b strings.Builder
	for i, s := range ss {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s.Content)
	}
	return b.String()
}

func normalizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func writeString(w io.Writer, s string) {
	_, _ = io.WriteString(w, s)
}

func xmlEscapeText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func xmlEscapeAttr(s string) string {
	s = xmlEscapeText(s)
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, `'`, "&apos;")
	return s
}
