package jsonrepair

import "strings"

// RepairInvalidStringEscapes removes invalid JSON escape markers inside string
// literals while preserving valid JSON escapes and text outside strings.
func RepairInvalidStringEscapes(jsonText string) (string, bool) {
	var b strings.Builder
	b.Grow(len(jsonText))
	inString := false
	changed := false

	for i := 0; i < len(jsonText); i++ {
		ch := jsonText[i]
		if !inString {
			b.WriteByte(ch)
			if ch == '"' {
				inString = true
			}
			continue
		}

		if ch == '"' {
			b.WriteByte(ch)
			inString = false
			continue
		}
		if ch != '\\' {
			b.WriteByte(ch)
			continue
		}
		if i+1 >= len(jsonText) {
			b.WriteByte(ch)
			continue
		}

		next := jsonText[i+1]
		if isValidEscape(next) {
			b.WriteByte(ch)
			b.WriteByte(next)
			i++
			continue
		}

		b.WriteByte(next)
		i++
		changed = true
	}

	if !changed {
		return jsonText, false
	}
	return b.String(), true
}

func isValidEscape(ch byte) bool {
	switch ch {
	case '"', '\\', '/', 'b', 'f', 'n', 'r', 't', 'u':
		return true
	default:
		return false
	}
}
