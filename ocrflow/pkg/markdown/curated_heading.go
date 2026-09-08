package markdown

import (
	"regexp"
	"strconv"
	"strings"
)

const CuratedHeadingPrefix = "curated-heading"

var curatedHeadingPattern = regexp.MustCompile(`^\[Curated heading level=([1-9][0-9]*): ([^\]]+)\]$`)

// ParseCuratedHeading parses a complete curated-heading annotation line.
func ParseCuratedHeading(line string) (int, string) {
	match := curatedHeadingPattern.FindStringSubmatch(strings.TrimSpace(line))
	if match == nil {
		return 0, ""
	}
	level, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, ""
	}
	return level, strings.TrimSpace(match[2])
}
