package markdown

import "regexp"

var emphasisPattern = regexp.MustCompile(`\*\*\*([^*]+)\*\*\*|\*\*([^*]+)\*\*|\*([^*]+)\*`)

func StripEmphasis(text string) string {
	return emphasisPattern.ReplaceAllString(text, "${1}${2}${3}")
}
