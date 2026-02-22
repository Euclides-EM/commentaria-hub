package formatcov

import (
	"bytes"
	"io"
	"sort"
	"strconv"
	"unicode/utf8"
)

func ConvertTXTToTEI(originalLines []string, linesByLang map[string][]string) ([]byte, error) {
	return ConvertTXTToTEIWithLang(originalLines, linesByLang, "und")
}

func ConvertTXTToTEIWithLang(originalLines []string, linesByLang map[string][]string, originalLang string) ([]byte, error) {
	if originalLang == "" {
		originalLang = "und"
	}
	var buf bytes.Buffer
	w := io.Writer(&buf)

	writeString(w, "<text>\n")
	writeString(w, "  <body>\n\n")

	// Original: one or more <p>; start a new <p> when a line begins with space or "¶"
	writeString(w, "    <p xml:lang=\""+xmlEscapeAttr(originalLang)+"\">\n")
	for i, line := range originalLines {
		if i > 0 && startsNewParagraph(line) {
			writeString(w, "    </p>\n")
			writeString(w, "    <p xml:lang=\""+xmlEscapeAttr(originalLang)+"\">\n")
		}
		id := lineID(i + 1)
		writeString(w, "      "+xmlEscapeText(line)+"<lb xml:id=\""+id+"\"/>\n")
	}
	writeString(w, "    </p>\n\n")

	// Translations: <div type="translations"> with one <ab> per language
	writeString(w, "    <div type=\"translations\">\n")
	langs := make([]string, 0, len(linesByLang))
	for lang := range linesByLang {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	for _, lang := range langs {
		lines := linesByLang[lang]
		writeString(w, "      <ab type=\"translation\" xml:lang=\""+xmlEscapeAttr(lang)+"\">\n")
		for i := range originalLines {
			segContent := ""
			if i < len(lines) {
				segContent = lines[i]
			}
			writeString(w, "        <seg corresp=\"#"+lineID(i+1)+"\">"+xmlEscapeText(segContent)+"</seg>\n")
		}
		writeString(w, "      </ab>\n\n")
	}
	writeString(w, "    </div>\n\n")
	writeString(w, "  </body>\n")
	writeString(w, "</text>")

	return buf.Bytes(), nil
}

func lineID(n int) string {
	return "l" + strconv.Itoa(n)
}

// startsNewParagraph reports whether the line begins with a space or "¶" (pilcrow).
func startsNewParagraph(line string) bool {
	r, _ := utf8.DecodeRuneInString(line)
	return r == ' ' || r == '¶'
}
