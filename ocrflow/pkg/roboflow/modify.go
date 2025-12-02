package roboflow

import (
	"encoding/json"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/coco"
	"os"
	"path/filepath"
)

func StretchCategoryTowardOtherCategory(dir string, apply *coco.StretchTowardsCategory) error {
	for _, subDir := range subDirs {
		fullDir := filepath.Join(dir, subDir)
		if _, err := os.Stat(filepath.Join(fullDir, "_annotations.coco.json")); os.IsNotExist(err) {
			continue
		}
		if err := modifyAnnotation(fullDir, apply); err != nil {
			return err
		}
	}
	return nil
}

func modifyAnnotation(dir string, apply *coco.StretchTowardsCategory) error {
	cocoPath := filepath.Join(dir, "_annotations.coco.json")

	data, err := os.ReadFile(cocoPath)
	if err != nil {
		return fmt.Errorf("read coco file: %w", err)
	}

	var c coco.Root
	if err := json.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("unmarshal coco: %w", err)
	}

	if err := coco.ApplyStretchTowardsCategory(&c, apply); err != nil {
		return fmt.Errorf("modify annotations: %w", err)
	}

	newData, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal coco: %w", err)
	}

	if err := os.WriteFile(cocoPath, newData, 0o644); err != nil {
		return fmt.Errorf("write coco file: %w", err)
	}

	return nil
}

func AddMargin(dir string, apply *coco.AddMargin) error {
	for _, subDir := range subDirs {
		fullDir := filepath.Join(dir, subDir)
		if err := modifyAddMargin(fullDir, apply); err != nil {
			return err
		}
	}
	return nil
}

func modifyAddMargin(dir string, apply *coco.AddMargin) error {
	cocoPath := filepath.Join(dir, "_annotations.coco.json")

	data, err := os.ReadFile(cocoPath)
	if err != nil {
		return fmt.Errorf("read coco file: %w", err)
	}

	var c coco.Root
	if err := json.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("unmarshal coco: %w", err)
	}

	if err := coco.ApplyAddMargin(&c, apply); err != nil {
		return fmt.Errorf("modify annotations: %w", err)
	}

	newData, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal coco: %w", err)
	}

	if err := os.WriteFile(cocoPath, newData, 0o644); err != nil {
		return fmt.Errorf("write coco file: %w", err)
	}

	return nil
}
