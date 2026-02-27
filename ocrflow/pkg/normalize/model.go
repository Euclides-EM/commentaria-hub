package normalize

import "regexp"

type MappedOriginal struct {
	Original string
	Mapped   string
}

type rule struct {
	re    *regexp.Regexp
	label string
}
