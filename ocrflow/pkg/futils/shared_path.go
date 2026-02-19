package futils

import (
	"path/filepath"
	"strings"
)

// SharedParent returns the deepest directory that contains both paths.
// Examples:
//
//	"/x/y/z/w", "/x/y/f/k" -> "/x/y"
//	"a/b/c", "a/b"         -> "a/b"
func SharedParent(p1, p2 string) string {
	p1 = filepath.Clean(p1)
	p2 = filepath.Clean(p2)

	// Decide whether to force absolute result or relative.
	abs1 := filepath.IsAbs(p1)
	abs2 := filepath.IsAbs(p2)

	parts1 := splitPath(p1)
	parts2 := splitPath(p2)

	n := min(len(parts1), len(parts2))
	i := 0
	for i < n && parts1[i] == parts2[i] {
		i++
	}

	if i == 0 {
		if abs1 && abs2 {
			return string(filepath.Separator) // only root in common
		}
		return "" // no shared parent for relative paths
	}

	common := filepath.Join(parts1[:i]...)
	if abs1 && abs2 {
		common = string(filepath.Separator) + common
	}
	return filepath.Clean(common)
}

func splitPath(p string) []string {
	p = strings.TrimPrefix(p, string(filepath.Separator))
	if p == "" {
		return nil
	}
	// filepath.Clean uses OS separators, so splitting by Separator is fine.
	return strings.Split(p, string(filepath.Separator))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
