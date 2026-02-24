package tei

import "strings"

func occKey(pageID, blockID, lineID string) string {
	return pageID + "||" + blockID + "||" + lineID
}

func surfaceID(pageID string) string {
	// Ensure valid xml:id
	return "page_" + sanitizeID(pageID)
}

func blockID(blockID string) string {
	return "b_" + sanitizeID(blockID)
}

func zoneID(kind, pageID, raw string) string {
	return "z_" + kind + "_" + sanitizeID(pageID) + "_" + sanitizeID(raw)
}

func sanitizeID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "x"
	}
	// conservative: keep letters, digits, underscore, dash, replace others with underscore
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}
