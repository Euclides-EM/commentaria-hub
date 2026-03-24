package textmatch

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

func FindLoosePhraseMatches(text string, featureValue string) [][2]int {
	normalized := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '-' {
			return -1
		}
		return r
	}, featureValue)
	if normalized == "" {
		return nil
	}

	runes := []rune(normalized)

	var pattern strings.Builder
	for i, r := range runes {
		if unicode.IsPunct(r) {
			pattern.WriteString(`[\p{P}\s-]+`)
		} else {
			pattern.WriteString(regexp.QuoteMeta(string(r)))
		}
		if i < len(runes)-1 {
			pattern.WriteString(`(?:[\s-])*`)
		}
	}

	rg, err := regexp.Compile("(?i)" + pattern.String())
	if err != nil {
		return nil
	}

	var matches [][2]int
	for from := 0; from < len(text); {
		loc := rg.FindStringIndex(text[from:])
		if loc == nil {
			break
		}

		start := from + loc[0]
		end := from + loc[1]
		matches = append(matches, [2]int{start, end})

		_, width := utf8.DecodeRuneInString(text[start:])
		if width <= 0 {
			width = 1
		}
		from = start + width
	}

	return matches
}
