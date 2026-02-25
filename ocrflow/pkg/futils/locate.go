package futils

import "os"

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
