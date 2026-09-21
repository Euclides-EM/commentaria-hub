package formatcov

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var yoloImageExtensions = []string{".jpg", ".jpeg", ".png", ".JPG", ".JPEG", ".PNG"}
var sourceImageExtensions = []string{".png", ".jpg", ".jpeg", ".tif", ".tiff", ".PNG", ".JPG", ".JPEG", ".TIF", ".TIFF"}

// FindYoloImage locates the image belonging to a YOLO label file. It supports
// both labels/images siblings and split labels with images at the dataset root.
func FindYoloImage(labelPath string) (string, error) {
	stem := strings.TrimSuffix(filepath.Base(labelPath), filepath.Ext(labelPath))
	labelDir := filepath.Dir(labelPath)
	imageDirs := []string{
		filepath.Clean(filepath.Join(labelDir, "..", "images")),
		filepath.Clean(filepath.Join(labelDir, "..", "..", "images")),
	}
	for _, imageDir := range imageDirs {
		for _, ext := range yoloImageExtensions {
			imagePath := filepath.Join(imageDir, stem+ext)
			if _, err := os.Stat(imagePath); err == nil {
				return imagePath, nil
			} else if !os.IsNotExist(err) {
				return "", err
			}
		}
	}
	return "", fmt.Errorf("cannot find an image for YOLO label %s", labelPath)
}

// ValidateYoloDataset verifies that a YOLO directory has usable class metadata
// and at least one self-contained label/image pair. Every discovered label must
// resolve to an image from the same YOLO bundle.
func ValidateYoloDataset(src string) error {
	labels, err := LoadYoloLabelmap(src)
	if err != nil {
		return err
	}
	if len(labels) == 0 {
		return fmt.Errorf("no YOLO classes found in %s", src)
	}

	pairs := 0
	for _, subDir := range SubDirs {
		labelDir := filepath.Join(src, subDir, "labels")
		entries, err := os.ReadDir(labelDir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read YOLO labels directory %s: %w", labelDir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".txt") || strings.EqualFold(entry.Name(), "labelmap.txt") {
				continue
			}
			labelPath := filepath.Join(labelDir, entry.Name())
			imagePath, err := FindYoloImage(labelPath)
			if err != nil {
				return err
			}
			relImagePath, err := filepath.Rel(src, imagePath)
			if err != nil {
				return err
			}
			if relImagePath == ".." || strings.HasPrefix(relImagePath, ".."+string(filepath.Separator)) {
				return fmt.Errorf("YOLO image %s for label %s is outside dataset %s", imagePath, labelPath, src)
			}
			pairs++
		}
	}
	if pairs == 0 {
		return fmt.Errorf("no YOLO label/image pairs found in %s", src)
	}
	return nil
}

// ValidateYoloImagesAgainstDataset ensures an imported YOLO bundle describes
// the same page geometry as the canonical dataset images. Exact dimensions are
// required because YOLO-to-ALTO writes pixel coordinates that are later used
// with those canonical images.
func ValidateYoloImagesAgainstDataset(src string, datasetImageDir string) error {
	if err := ValidateYoloDataset(src); err != nil {
		return err
	}
	for _, subDir := range SubDirs {
		labelDir := filepath.Join(src, subDir, "labels")
		entries, err := os.ReadDir(labelDir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read YOLO labels directory %s: %w", labelDir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".txt") || strings.EqualFold(entry.Name(), "labelmap.txt") {
				continue
			}
			labelPath := filepath.Join(labelDir, entry.Name())
			yoloImagePath, err := FindYoloImage(labelPath)
			if err != nil {
				return err
			}
			labelStem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			pageStem, ok := CutPagePrefix(labelStem)
			if !ok {
				return fmt.Errorf("YOLO label %s does not identify a page-NNNN dataset image", labelPath)
			}
			datasetImagePath, err := findImageByStem(datasetImageDir, pageStem)
			if err != nil {
				return fmt.Errorf("match YOLO label %s to canonical dataset image: %w", labelPath, err)
			}
			yoloWidth, yoloHeight, err := imageSize(yoloImagePath)
			if err != nil {
				return fmt.Errorf("read YOLO image %s: %w", yoloImagePath, err)
			}
			datasetWidth, datasetHeight, err := imageSize(datasetImagePath)
			if err != nil {
				return fmt.Errorf("read canonical dataset image %s: %w", datasetImagePath, err)
			}
			if yoloWidth != datasetWidth || yoloHeight != datasetHeight {
				return fmt.Errorf(
					"YOLO image %s is %dx%d but canonical dataset image %s is %dx%d; geometrically transformed YOLO imports are not supported",
					yoloImagePath, yoloWidth, yoloHeight, datasetImagePath, datasetWidth, datasetHeight,
				)
			}
		}
	}
	return nil
}

func findImageByStem(dir string, stem string) (string, error) {
	for _, ext := range sourceImageExtensions {
		imagePath := filepath.Join(dir, stem+ext)
		if _, err := os.Stat(imagePath); err == nil {
			return imagePath, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("no image named %s with a supported extension in %s", stem, dir)
}
