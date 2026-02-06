package store

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
)

func InferPages(dstPath string, format ocrflow.AnnotationFormat) ([]int, error) {
	switch format {
	case ocrflow.AnnotationFormatAlto:
		return inferPagesFromAltoDir(dstPath)
	case ocrflow.AnnotationFormatYolo:
		return inferPagesFromYoloDir(dstPath)
	default:
		return inferPagesFromImgDir(dstPath)
	}
}

func inferPagesFromImgDir(path string) ([]int, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	res := make([]int, 0)
	for _, f := range files {
		if f.IsDir() && f.Name() == ".DS_Store" {
			continue
		}
		pageNum, err := pagesparser.FileNameToPage(f.Name())
		if err != nil {
			log.Printf("could not parse file %s, expected format page-0001.<ext>, skipping: %v", f.Name(), err)
			continue
		}
		res = append(res, pageNum)
	}
	return res, nil
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

		re := regexp.MustCompile(`^page-(\d{4})`)
		m := re.FindStringSubmatch(f.Name())
		if len(m) != 2 {
			log.Printf("could not parse file %s, expected format page-0001.txt, skipping", f.Name())
			continue
		}
		pageNum, err = strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse file name %s: %w", f.Name(), err)
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
		if f.IsDir() || f.Name() == "METS.xml" {
			continue
		}
		pageNum, err := pagesparser.FileNameToPage(f.Name())
		if err != nil {
			log.Printf("could not parse file %s, expected format page-0001.xml, skipping: %v", f.Name(), err)
			continue
		}
		res = append(res, pageNum)
	}
	return res, nil
}
