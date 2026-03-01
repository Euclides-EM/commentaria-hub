package name

import (
	"fmt"
)

func NextAvailable(existing []string, base string) string {
	set := make(map[string]struct{}, len(existing))
	for _, s := range existing {
		set[s] = struct{}{}
	}

	// Try base first
	if _, ok := set[base]; !ok {
		return base
	}

	// Then base 1, base 2, ...
	for i := 1; ; i++ {
		cand := fmt.Sprintf("%s %d", base, i)
		if _, ok := set[cand]; !ok {
			return cand
		}
	}
}
