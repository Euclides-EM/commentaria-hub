package normalize

import (
	"regexp"
	"strings"

	"github.com/samber/lo"
)

func byRegex(rules []rule, defaultVal string) func(string) string {
	return func(s string) string {
		result := make([]string, 0)
		norm := String(s)
		if norm == "" {
			return defaultVal
		}

		for _, r := range rules {
			if r.re.MatchString(norm) {
				result = append(result, r.label)
			}
		}

		result = lo.Uniq(result)
		if len(result) == 0 {
			return defaultVal
		}

		return strings.Join(result, "::")
	}
}

type rule struct {
	re    *regexp.Regexp
	label string
}

//var splitRe = regexp.MustCompile(`, | et | en | & `)
