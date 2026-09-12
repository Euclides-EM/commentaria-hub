package markdown

import (
	"regexp"
	"strconv"
	"strings"
)

const CuratedHeadingPrefix = "curated-heading"
const DefaultIndexType = "default"

var (
	curatedHeadingPattern = regexp.MustCompile(`^\[Curated heading ([^:]+): ([^\]]+)\]$`)
	indexTypePattern      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	indexLevelPattern     = regexp.MustCompile(`^[1-9][0-9]*$`)
)

func IsValidIndexType(indexType string) bool {
	return indexTypePattern.MatchString(indexType)
}

type CuratedHeading struct {
	Level int
	Type  string
	Text  string
}

// ParseTypedCuratedHeading parses a complete curated-heading annotation line.
// An omitted type belongs to the default index layer. The name "default" is
// reserved for that layer and therefore cannot be used explicitly.
func ParseTypedCuratedHeading(line string) (CuratedHeading, bool) {
	match := curatedHeadingPattern.FindStringSubmatch(strings.TrimSpace(line))
	if match == nil {
		return CuratedHeading{}, false
	}

	level := 0
	headingType := DefaultIndexType
	typeSeen := false
	for _, attribute := range strings.Fields(match[1]) {
		name, value, ok := strings.Cut(attribute, "=")
		if !ok || value == "" {
			return CuratedHeading{}, false
		}
		switch name {
		case "level":
			if level != 0 || !indexLevelPattern.MatchString(value) {
				return CuratedHeading{}, false
			}
			parsedLevel, err := strconv.Atoi(value)
			if err != nil || parsedLevel < 1 {
				return CuratedHeading{}, false
			}
			level = parsedLevel
		case "type":
			if typeSeen || value == DefaultIndexType || !IsValidIndexType(value) {
				return CuratedHeading{}, false
			}
			headingType = value
			typeSeen = true
		default:
			return CuratedHeading{}, false
		}
	}
	if level == 0 {
		return CuratedHeading{}, false
	}
	return CuratedHeading{Level: level, Type: headingType, Text: strings.TrimSpace(match[2])}, true
}

// ParseCuratedHeading parses a complete curated-heading annotation line.
func ParseCuratedHeading(line string) (int, string) {
	heading, ok := ParseTypedCuratedHeading(line)
	if !ok {
		return 0, ""
	}
	return heading.Level, heading.Text
}
