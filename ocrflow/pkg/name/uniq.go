package name

import (
	"fmt"

	"github.com/samber/lo"
)

func NextAvailable(existing []string, name string) string {
	return nextAvailable(existing, name, 0)
}

func nextAvailable(existing []string, name string, i int) string {
	if len(lo.Filter(existing, func(d string, _ int) bool { return d == name })) == 0 {
		if i == 0 {
			return name
		}
		return fmt.Sprintf("%s %d", name, i)
	}
	return nextAvailable(existing, name, i+1)
}
