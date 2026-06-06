package diagramcrops

import (
	"regexp"
	"strings"
)

var validKeyRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var volSuffixRe = regexp.MustCompile(`^(.+)_vol([0-9]+)$`)

func ValidKey(key string) bool {
	return key != "" && key != "." && !strings.HasPrefix(key, ".") && validKeyRe.MatchString(key)
}

func BaseKey(key string) string {
	if matches := volSuffixRe.FindStringSubmatch(key); len(matches) == 3 {
		return matches[1]
	}
	return key
}
