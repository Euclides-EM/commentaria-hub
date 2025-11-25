package krakenwrapper

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// --- CONFIGURATION ---
const (
	//krakenModel = "catmus-print-fondue-large.mlmodel"
	// todo: Kraken is the command-line tool to execute. Ensure it's in your system's PATH.
	// $ python3 -m venv kraken_env
	//$ source kraken_env/bin/activate
	//
	//# 2. Install kraken
	//$ pip install kraken
	//
	//# (Optional) Install PDF/multi-image TIFF support
	//$ pip install kraken[pdf]
	// kraken get 10.5281/zenodo.10592716
	krakenCommand = "kraken"
)

var imageFormats = []string{".tif", ".tiff", ".png"}

func Recognize(inputDir, outputDir, krakenModel string, filenames []string) error {
	if len(filenames) == 0 {
		log.Println("No filenames provided for OCR processing.")
		return nil
	}
	if _, err := os.Stat(inputDir); err != nil {
		return fmt.Errorf("input directory does not exist: %w", err)
	}
	files := make([]string, len(filenames))
	for _, filename := range filenames {
		if !slices.Contains(imageFormats, strings.ToLower(filepath.Ext(filename))) {
			return fmt.Errorf("input file %s is not a supported image format (TIFF/PNG)", filename)
		}
		inputPath := filepath.Join(inputDir, filename)
		if _, err := os.Stat(inputPath); err != nil {
			return fmt.Errorf("input file %s does not exist: %w", inputPath, err)
		}
		if strings.ToLower(filepath.Ext(filename)) != ".tif" && strings.ToLower(filepath.Ext(filename)) != ".tiff" && strings.ToLower(filepath.Ext(filename)) != ".png" {
			return fmt.Errorf("input file %s is not a supported image format (TIFF/PNG)", filename)
		}
		files = append(files, inputPath)
	}
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("could not clean old output directory: %w", err)
	}
	if err := os.Mkdir(outputDir, 0755); err != nil {
		return fmt.Errorf("could not create output directory: %w", err)
	}

	return processImages(files, outputDir, krakenModel)
}

// processImages iterates over the list of images and runs the Kraken command for each.
func processImages(images []string, outputDir, krakenModel string) error {
	for i, inputPath := range images {
		base := filepath.Base(inputPath)
		nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
		outputPath := filepath.Join(outputDir, nameWithoutExt+".txt")

		log.Printf("[%d/%d] Running OCR on: %s", i+1, len(images), inputPath)

		args := []string{
			"-i", inputPath,
			outputPath,
			"binarize",
			"segment", "-bl",
			"ocr",
			"-m", krakenModel,
		}

		cmd := exec.Command(krakenCommand, args...)

		output, err := cmd.CombinedOutput()

		if err != nil {
			log.Printf("!!! ERROR processing %s:", inputPath)
			log.Printf(string(output))
			return fmt.Errorf("kraken command failed for %s: %w", inputPath, err)
		}

		log.Printf("   -> Successfully saved result to: %s", outputPath)
	}
	return nil
}
