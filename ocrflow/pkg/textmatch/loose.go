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

func FindFuzzyPhraseMatch(text string, featureValue string, maxEdits int) ([2]int, bool) {
	if maxEdits < 0 {
		return [2]int{}, false
	}

	normalizedValue, _, _ := normalizeForFuzzyMatch(featureValue)
	if len(normalizedValue) < 12 {
		return [2]int{}, false
	}

	normalizedText, starts, ends := normalizeForFuzzyMatch(text)
	if len(normalizedText) < len(normalizedValue)-maxEdits {
		return [2]int{}, false
	}

	bestDistance := maxEdits + 1
	bestLengthDelta := maxEdits + 1
	best := [2]int{}
	bestFound := false
	for start := range normalizedText {
		minEnd := start + len(normalizedValue) - maxEdits
		maxEnd := start + len(normalizedValue) + maxEdits
		if minEnd < start+1 {
			minEnd = start + 1
		}
		if maxEnd > len(normalizedText) {
			maxEnd = len(normalizedText)
		}
		for end := minEnd; end <= maxEnd; end++ {
			distance := boundedEditDistance(normalizedValue, normalizedText[start:end], maxEdits)
			if distance > maxEdits || distance > bestDistance {
				continue
			}
			lengthDelta := absInt((end - start) - len(normalizedValue))
			span := [2]int{starts[start], ends[end-1]}
			if distance < bestDistance || !bestFound || lengthDelta < bestLengthDelta || (lengthDelta == bestLengthDelta && span[1]-span[0] < best[1]-best[0]) {
				bestDistance = distance
				bestLengthDelta = lengthDelta
				best = span
				bestFound = true
			}
		}
	}

	return best, bestFound
}

func normalizeForFuzzyMatch(s string) ([]rune, []int, []int) {
	var normalized []rune
	var starts []int
	var ends []int
	for byteStart, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			continue
		}
		normalized = append(normalized, unicode.ToLower(r))
		starts = append(starts, byteStart)
		ends = append(ends, byteStart+utf8.RuneLen(r))
	}
	return normalized, starts, ends
}

func boundedEditDistance(a, b []rune, maxDistance int) int {
	if len(a)-len(b) > maxDistance || len(b)-len(a) > maxDistance {
		return maxDistance + 1
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		rowMin := curr[0]
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
			if curr[j] < rowMin {
				rowMin = curr[j]
			}
		}
		if rowMin > maxDistance {
			return maxDistance + 1
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
