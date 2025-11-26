package krakenwrapper

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var imageFormats = []string{".tif", ".tiff", ".png"}

func Recognize(inputDir, outputDir, segmentationModel, ocrModel string, filenames []string) (<-chan error, error) {
	if len(filenames) == 0 {
		log.Println("No filenames provided for OCR processing.")
		ch := make(chan error, 1)
		ch <- nil
		close(ch)
		return ch, nil
	}

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

	errCh := processImages(files, outputDir, segmentationModel, ocrModel)
	return errCh, nil
}

func processImages(images []string, outputDir, segmentationModel, ocrModel string) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- runProcessImages(images, outputDir, segmentationModel, ocrModel)
	}()

	return errCh
}

func runProcessImages(images []string, outputDir, segmentationModel, ocrModel string) error {
	err, inputOutputPairs, err2 := toInputOutputPairs(images, outputDir)
	if err2 != nil {
		return err2
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

	args = append(args,
		"segment",
		"--yolo", segmentationModel,
		"ocr",
		"--model", ocrModel)

	cmd := exec.Command("yaltai", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start yaltai: %w", err)
	}

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Printf("[yaltai stdout] %s", scanner.Text())
		}
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[yaltai stderr] %s", scanner.Text())
		}
	}()

	// Wait for exit
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("yaltai command failed: %w", err)
	}

	log.Printf("Yaltai command succeeded. Args: %v", args)
	return nil
}

// toInputOutputPairs converts a list of image paths to input-output pairs for processing.
// It is used to prepare the arguments for the Kraken command.
func toInputOutputPairs(images []string, outputDir string) (error, [][2]string, error) {
	outputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, nil, fmt.Errorf("could not determine absolute path of output directory: %w", err)
	}
	inputOutputPairs := make([][2]string, 0, len(images))
	for _, inputPath := range images {
		inputPath, err = filepath.Abs(inputPath)
		if err != nil {
			return nil, nil, fmt.Errorf("could not determine absolute path of input image %s: %w", inputPath, err)
		}
		base := filepath.Base(inputPath)
		nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
		outputPath := filepath.Join(outputDir, nameWithoutExt+".xml")

		inputOutputPairs = append(inputOutputPairs, [2]string{inputPath, outputPath})
	}
	return err, inputOutputPairs, nil
}
