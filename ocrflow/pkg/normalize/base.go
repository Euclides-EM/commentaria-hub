package normalize

import (
	"regexp"
	"strings"
)

func byRegex(rules []rule, defaultVal string) func(string) string {
	return func(s string) string {
		parts := splitRe.Split(s, -1)

		var result []string
		for _, input := range parts {
			norm := String(input)
			if norm == "" {
				continue
			}

			matched := defaultVal
			for _, r := range rules {
				if r.re.MatchString(norm) {
					matched = r.label
					break
				}
			}

			result = append(result, matched)
		}

		return strings.Join(result, "::")
	}
}

type rule struct {
	re    *regexp.Regexp
	label string
}

var splitRe = regexp.MustCompile(`, | et | en | & `)
