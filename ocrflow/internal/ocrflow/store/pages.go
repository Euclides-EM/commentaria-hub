package store

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"log"
	"os"
	"path/filepath"
)

func InferPages(dstPath string, format model.AnnotationFormat) ([]int, error) {
	switch format {
	case model.AnnotationFormatAlto:
		return inferPagesFromAltoDir(dstPath)
	case model.AnnotationFormatYolo:
		return inferPagesFromYoloDir(dstPath)
	default:
		return nil, nil
	}
}

func inferPagesFromYoloDir(path string) ([]int, error) {
	rootFiles, err := inferPagesFromYoloInnerDir(path)
	if err != nil {
		return nil, err
	}
	if len(rootFiles) > 0 {
		return rootFiles, nil
	}

	subdirs, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	res := make([]int, 0)
	for _, d := range subdirs {
		if !d.IsDir() {
			continue
		}
		subdirPages, err := inferPagesFromYoloInnerDir(filepath.Join(path, d.Name()))
		if err != nil {
			return nil, err
		}
		res = append(res, subdirPages...)
	}
	return res, nil
}

func inferPagesFromYoloInnerDir(path string) ([]int, error) {
	res := make([]int, 0)
	files, err := os.ReadDir(filepath.Join(path, "labels"))
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, f := range files {
		var pageNum int
		_, err := fmt.Sscanf(f.Name(), "page-%04d.txt", &pageNum)
		if err != nil {
			log.Printf("could not parse file %s, expected format page-0001.txt, skipping: %v", f.Name(), err)
			continue
		}
		res = append(res, pageNum)
	}
	return res, nil
}

func inferPagesFromAltoDir(p string) ([]int, error) {
	files, err := os.ReadDir(p)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	res := make([]int, 0)
	for _, f := range files {
		var pageNum int
		_, err := fmt.Sscanf(f.Name(), "page-%04d.xml", &pageNum)
		if err != nil {
			log.Printf("could not parse file %s, expected format page-0001.xml, skipping: %v", f.Name(), err)
			continue
		}
		res = append(res, pageNum)
	}
	return res, nil
}
