package krakenwrapper

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var imageFormats = []string{".tif", ".tiff", ".png"}

func Recognize(inputDir, outputDir, segmentationModel string, filenames []string) (<-chan error, error) {
	if _, err := os.Stat(inputDir); err != nil {
		return nil, fmt.Errorf("input directory does not exist: %w", err)
	}

	files := make([]string, 0)
	for _, filename := range filenames {
		ext := strings.ToLower(filepath.Ext(filename))
		if !slices.Contains(imageFormats, ext) {
			return nil, fmt.Errorf("input file %s is not a supported image format (TIFF/PNG)", filename)
		}
		inputPath := filepath.Join(inputDir, filename)
		if _, err := os.Stat(inputPath); err != nil {
			return nil, fmt.Errorf("input file %s does not exist: %w", inputPath, err)
		}
		files = append(files, inputPath)
	}

	if err := os.RemoveAll(outputDir); err != nil {
		return nil, fmt.Errorf("could not clean old output directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create output directory: %w", err)
	}

	errCh := processImages(files, outputDir, segmentationModel)
	return errCh, nil
}

func processImages(images []string, outputDir, segmentationModel string) <-chan error {

	// todo: as opposed to the flow that converts from yolo to alto, here we do not copy the images to the output dir
	//  Moreover, the filename in the ALTO result might match the input image full path,
	//  which can be problematic later on, for example, in the Escriptorium upload.

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- runProcessImages(images, outputDir, segmentationModel)
	}()

	return errCh
}

func runProcessImages(images []string, outputDir, segmentationModel string) error {
	inputOutputPairs, err := toInputOutputPairs(images, outputDir)
	if err != nil {
		return err
	}
	args := []string{
		"kraken",
		"--device", "cpu",
		"--alto",
	}

	for _, pair := range inputOutputPairs {
		inputPath := pair[0]
		outputPath := pair[1]
		args = append(args, "-i", inputPath, outputPath)
	}

	args = append(args, "segment", "--yolo", segmentationModel)

	//if ocrModel != "" {
	//	args = append(args, "ocr", "--model", ocrModel)
	//}

	if err := envexec.Cmd("yaltai", args...); err != nil {
		return fmt.Errorf("kraken recognition failed: %w", err)
	}

	// go over all output files and fix the fileName element in the ALTO XML
	for _, pair := range inputOutputPairs {
		outputPath := pair[1]
		if err := FixAltoFileName(outputPath, outputPath); err != nil {
			return fmt.Errorf("could not fix ALTO file name in %s: %w", outputPath, err)
		}
	}
	return nil
}

// toInputOutputPairs converts a list of image paths to input-output pairs for processing.
// It is used to prepare the arguments for the Kraken command.
func toInputOutputPairs(images []string, outputDir string) ([][2]string, error) {
	outputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("could not determine absolute path of output directory: %w", err)
	}
	inputOutputPairs := make([][2]string, 0, len(images))
	for _, inputPath := range images {
		inputPath, err = filepath.Abs(inputPath)
		if err != nil {
			return nil, fmt.Errorf("could not determine absolute path of input image %s: %w", inputPath, err)
		}
		base := filepath.Base(inputPath)
		nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
		outputPath := filepath.Join(outputDir, nameWithoutExt+".xml")

		inputOutputPairs = append(inputOutputPairs, [2]string{inputPath, outputPath})
	}
	return inputOutputPairs, nil
}
