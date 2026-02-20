package futils

import (
	"path/filepath"
	"strings"
)

func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".tif" || ext == ".tiff" || ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}
