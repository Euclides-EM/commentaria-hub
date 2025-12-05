package roboflow

import (
	"encoding/json"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/coco"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"os"
	"path/filepath"
)

var subDirs = []string{"", "train", "valid", "test"}

func SlicePages(dir string, pages []int) error {
	for _, subDir := range subDirs {
		fullDir := filepath.Join(dir, subDir)
		if _, err := os.Stat(filepath.Join(fullDir, "_annotations.coco.json")); os.IsNotExist(err) {
			continue
		}
		if err := slicePages(fullDir, pages); err != nil {
			return fmt.Errorf("slice pages in %s: %w", fullDir, err)
		}
	}
	return nil
}

func slicePages(dir string, pages []int) error {
	cocoPath := filepath.Join(dir, "_annotations.coco.json")

	data, err := os.ReadFile(cocoPath)
	if err != nil {
		return fmt.Errorf("read coco file: %w", err)
	}

	var c coco.Root
	if err := json.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("unmarshal coco: %w", err)
	}

	// build set of filenames to keep
	keepFileNames := make(map[string]struct{}, len(pages))
	for _, p := range pages {
		name := pagesparser.PageToPNGFilename(p)
		keepFileNames[name] = struct{}{}
	}

	// filter images in COCO by filename
	newImages := make([]coco.Image, 0, len(c.Images))
	keepImageIDs := make(map[int]struct{})

	for _, img := range c.Images {
		if _, ok := keepFileNames[img.FileName]; ok {
			newImages = append(newImages, img)
			keepImageIDs[img.ID] = struct{}{}
		}
	}

	// filter annotations by kept image ids
	newAnnotations := make([]coco.Annotation, 0, len(c.Annotations))
	for _, ann := range c.Annotations {
		if _, ok := keepImageIDs[ann.ImageID]; ok {
			newAnnotations = append(newAnnotations, ann)
		}
	}

	c.Images = newImages
	c.Annotations = newAnnotations

	// delete PNGs on disk that are not in keepFileNames
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "_annotations.coco.json" {
			continue
		}
		if filepath.Ext(name) != ".png" {
			continue
		}
		if _, ok := keepFileNames[name]; !ok {
			if err := os.Remove(filepath.Join(dir, name)); err != nil {
				return fmt.Errorf("remove %s: %w", name, err)
			}
		}
	}

	out, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal coco: %w", err)
	}

	if err := os.WriteFile(cocoPath, out, 0o644); err != nil {
		return fmt.Errorf("write coco file: %w", err)
	}

	return nil
}
