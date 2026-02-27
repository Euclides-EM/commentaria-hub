package normalize

import (
	"strings"

	"github.com/samber/lo"
)

func byRegex(rules []rule) func(string) []MappedOriginal {
	return func(s string) []MappedOriginal {
		norm := normalizeString(s)
		if norm == "" {
			return nil
		}

		var out []MappedOriginal

		for _, r := range rules {
			matches := r.re.FindAllString(norm, -1)
			for _, m := range matches {
				m = strings.TrimSpace(m)
				if m == "" {
					continue
				}
				out = append(out, MappedOriginal{
					Original: m,
					Mapped:   r.label,
				})
			}
		}

		// dedupe exact pairs, keep stable order
		out = lo.UniqBy(out, func(x MappedOriginal) string {
			return x.Original + "\x00" + x.Mapped
		})

		return out
	}
}
