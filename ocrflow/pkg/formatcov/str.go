package formatcov

import (
	"strings"

	"github.com/samber/lo"
)

func SplitLines(s string) []string {
	lines := make([]string, 0)
	for _, line := range lo.Filter(strings.Split(s, "\n"), func(line string, _ int) bool {
		return strings.TrimSpace(line) != ""
	}) {
		lines = append(lines, line)
	}
	return lines
}
