package futils

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func FindFileByExtension(dir string, exts ...string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		if slices.Contains(exts, ext) {
			return filepath.Join(dir, name), nil
		}
	}

	return "", fmt.Errorf("no yml file found in %s", dir)
}
