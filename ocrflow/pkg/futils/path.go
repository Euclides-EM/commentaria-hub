package futils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SafeJoin joins destRoot and relPath into a single path, ensuring that the resulting path is within destRoot.
func SafeJoin(destRoot, relPath string) (string, error) {
	rootAbs, err := filepath.Abs(destRoot)
	if err != nil {
		return "", err
	}
	dest := filepath.Join(rootAbs, filepath.FromSlash(relPath))
	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return "", err
	}
	// ensure dest is under root
	if destAbs != rootAbs && !strings.HasPrefix(destAbs, rootAbs+string(filepath.Separator)) {
		return "", fmt.Errorf("refusing to write outside destination: %s", destAbs)
	}
	return destAbs, nil
}

func SafeBase(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	base := filepath.Base(v)
	if base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}
