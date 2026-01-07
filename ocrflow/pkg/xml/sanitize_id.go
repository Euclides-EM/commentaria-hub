package xml

import "strings"

func SanitizeXMLID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))

	for i, r := range s {
		ok := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			r == '_' ||
			(i > 0 && r >= '0' && r <= '9') ||
			(i > 0 && r == '-' || r == '.')

		if ok {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}

	out := b.String()
	// xml:id cannot start with a digit or punctuation, prefix if needed
	if out == "" {
		return ""
	}
	first := out[0]
	if (first >= '0' && first <= '9') || first == '-' || first == '.' {
		out = "l_" + out
	}
	return out
}
