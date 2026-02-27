package normalize

import (
	"regexp"
)

var rulesInstitutions = []rule{
	{
		re:    regexp.MustCompile(`\b(?:(?:la\s*)?(?:compagnie|compa[nñ]i[aí]a|compania))\s+de\s+(?:jesus|iesvs|jesvs)|\b(?:soc\.?|soci[eé]t\.?|societate|societ\.)\s*(?:jesu|iesv|jesv)|\b(?:societatis)(?:\s+(?:jesu|iesv))?(?:\s+gymnasio)?\b|\bsociety of jesus\b|\bjesuite\b|\bpanormitano.*sicili\b|\bherbipolitano.*franconi\b|\bgymnasio.*(?:jesu|iesv|jesv)\b`),
		label: "Jesuits",
	},
}

func Institution(institutions string) []MappedOriginal {
	return byRegex(rulesInstitutions)(institutions)
}
