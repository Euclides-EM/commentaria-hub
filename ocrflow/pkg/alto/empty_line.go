package alto

import "github.com/samber/lo"

func IsEmptyLine(l TextLine) bool {
	return len(l.Strings) == 0 || lo.EveryBy(l.Strings, func(s AltoString) bool {
		return s.Content == ""
	})
}
