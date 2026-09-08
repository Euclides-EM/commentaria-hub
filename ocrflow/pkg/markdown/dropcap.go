package markdown

import (
	"regexp"
	"strings"
)

var dropcapPattern = regexp.MustCompile(`^\{dropcap:([^|}]+)\|lines=([^|}]+)\|style=(plain|decorated|unknown)(?:\|decoration="([^"]*)")?\}`)

// Dropcap is a parsed drop-cap annotation from the transcription Markdown
// dialect.
type Dropcap struct {
	Text       string
	Lines      string
	Style      string
	Decoration string
}

// ParseDropcapPrefix parses a drop-cap annotation at the start of text. The
// returned remainder begins immediately after the annotation.
func ParseDropcapPrefix(text string) (dropcap Dropcap, remainder string, ok bool) {
	match := dropcapPattern.FindStringSubmatch(text)
	if match == nil {
		return Dropcap{}, text, false
	}
	return Dropcap{
		Text:       match[1],
		Lines:      match[2],
		Style:      match[3],
		Decoration: match[4],
	}, text[len(match[0]):], true
}

// ExpandDropcaps projects drop-cap annotations to their visible text while
// leaving all other Markdown unchanged.
func ExpandDropcaps(text string) string {
	var result strings.Builder
	for len(text) > 0 {
		start := strings.Index(text, "{dropcap:")
		if start < 0 {
			result.WriteString(text)
			break
		}
		result.WriteString(text[:start])
		text = text[start:]
		dropcap, remainder, ok := ParseDropcapPrefix(text)
		if !ok {
			result.WriteByte(text[0])
			text = text[1:]
			continue
		}
		result.WriteString(dropcap.Text)
		text = remainder
	}
	return result.String()
}
