package futils

import (
	"os"
	"path/filepath"
	"strings"
)

func LocateFileInDir(dir string, filter func(filename string) bool) string {
	de, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, entry := range de {
		if entry.IsDir() {
			continue
		}
		if filter(entry.Name()) {
			return dir + "/" + entry.Name()
		}
	}
	return ""
}

func DirIncludesByFileExt(dir, ext string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), "."+ext) {
			return true, nil
		}
	}
	return false, nil
}
