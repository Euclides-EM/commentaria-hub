package alto

import "strings"

func ExtractTextFromLine(l TextLine) string {
	if len(l.Strings) == 0 {
		return ""
	}
	var b strings.Builder
	for i := range l.Strings {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(l.Strings[i].Content)
	}
	return b.String()
}

func ExtractTextContentFromBlock(b *TextBlock) string {
	lines := ExtractTextContentsFromBlock(b)
	combined := ""
	for _, content := range lines {
		c := strings.TrimSpace(content)
		if strings.HasSuffix(c, "¬") {
			combined += strings.TrimSuffix(c, "¬")
		} else {
			combined += c + " "
		}
	}
	combined = strings.TrimSpace(combined)
	return combined
}
