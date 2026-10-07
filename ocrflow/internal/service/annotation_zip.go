package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (a *Annotation) Zip(datasetID, id string, includeAlto, includeImages bool) (string, func(io.Writer) error, error) {
	if !includeAlto && !includeImages {
		return "", nil, fmt.Errorf("nothing to include in annotation zip, select ALTO and/or images")
	}
	ann, err := a.Get(datasetID, id)
	if err != nil {
		return "", nil, err
	}
	altoDir := a.fileSysMgt.DatasetAnnotationAltoDir(ann)
	imagesDir := a.fileSysMgt.DatasetImagesDirByID(datasetID)
	var imageNames []string
	if includeImages {
		entries, err := os.ReadDir(imagesDir)
		if err != nil && !os.IsNotExist(err) {
			return "", nil, fmt.Errorf("read dataset images dir: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				imageNames = append(imageNames, entry.Name())
			}
		}
	}
	write := func(w io.Writer) error {
		zw := zip.NewWriter(w)
		if includeAlto {
			if err := addDirToZip(zw, altoDir, "alto"); err != nil {
				return err
			}
		}
		for _, imageName := range imageNames {
			if err := addFileToZip(zw, filepath.Join(imagesDir, imageName), "imgs/"+imageName); err != nil {
				return err
			}
		}
		return zw.Close()
	}
	name := strings.NewReplacer("/", "_", "\\", "_", "\"", "_").Replace(strings.TrimSpace(ann.Name))
	if name == "" {
		name = ann.ID
	}
	return name + ".zip", write, nil
}
