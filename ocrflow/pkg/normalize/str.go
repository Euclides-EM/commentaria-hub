package normalize

import (
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

func String(s string) string {
	s = strings.ReplaceAll(s, "¬", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "’", "'")
	s = strings.ReplaceAll(s, "\u017F", "s") // long s -> s
	s = strings.TrimSpace(strings.ToLower(s))

	// Unicode NFD decomposition
	t := norm.NFD.String(s)

	// remove diacritic marks
	var b strings.Builder
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}

	return b.String()
}
