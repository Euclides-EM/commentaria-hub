package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func discoverImages(root string) ([]imageInput, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	var images []imageInput
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !isImagePath(path) {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(relative), "/")
		if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" {
			return fmt.Errorf("image %s must be inside a volume directory", path)
		}
		images = append(images, imageInput{Path: path, Volume: parts[0]})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("no supported images found in %s", root)
	}
	sort.Slice(images, func(i, j int) bool { return images[i].Path < images[j].Path })
	return images, nil
}

func isImagePath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
