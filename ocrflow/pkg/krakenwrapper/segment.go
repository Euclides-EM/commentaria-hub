package krakenwrapper

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
)

var imageFormats = []string{".tif", ".tiff", ".png"}

func Segment(inputDir, outputDir, segmentationModel string, filenames []string) (<-chan error, error) {
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
	//  which can be problematic later on, for example, in the eScriptorium upload.

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- runProcessImages(images, outputDir, segmentationModel)
	}()

	return errCh
}

// maxParallelKraken caps the number of concurrent Kraken processes (memory-heavy).
const maxParallelKraken = 8

func runProcessImages(images []string, outputDir, segmentationModel string) error {
	inputOutputPairs, err := toInputOutputPairs(images, outputDir)
	if err != nil {
		return err
	}
	if len(inputOutputPairs) == 0 {
		return nil
	}

	workers := runtime.NumCPU()
	if workers > maxParallelKraken {
		workers = maxParallelKraken
	}
	if workers > len(inputOutputPairs) {
		workers = len(inputOutputPairs)
	}
	if workers < 1 {
		workers = 1
	}

	chunks := chunkPairs(inputOutputPairs, workers)
	var firstErr error
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(chunks))
	for _, chunk := range chunks {
		chunk := chunk
		go func() {
			defer wg.Done()
			if err := runKrakenSegment(chunk, segmentationModel); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}

	// Fix fileName in ALTO XML for all outputs (parallel by file).
	var fixWg sync.WaitGroup
	var fixMu sync.Mutex
	var fixErr error
	fixWg.Add(len(inputOutputPairs))
	for _, pair := range inputOutputPairs {
		outputPath := pair[1]
		go func() {
			defer fixWg.Done()
			if err := RemovePathFromAltoImgFileName(outputPath, outputPath); err != nil {
				fixMu.Lock()
				if fixErr == nil {
					fixErr = fmt.Errorf("could not fix ALTO file name in %s: %w", outputPath, err)
				}
				fixMu.Unlock()
			}
		}()
	}
	fixWg.Wait()
	if fixErr != nil {
		return fixErr
	}
	return nil
}

// chunkPairs splits pairs into n roughly equal chunks for parallel workers.
func chunkPairs(pairs [][2]string, n int) [][][2]string {
	if n <= 0 || len(pairs) == 0 {
		return nil
	}
	if n >= len(pairs) {
		chunks := make([][][2]string, len(pairs))
		for i := range pairs {
			chunks[i] = [][2]string{pairs[i]}
		}
		return chunks
	}
	chunks := make([][][2]string, n)
	base, extra := len(pairs)/n, len(pairs)%n
	idx := 0
	for i := 0; i < n; i++ {
		size := base
		if i < extra {
			size++
		}
		chunks[i] = pairs[idx : idx+size]
		idx += size
	}
	return chunks
}

func runKrakenSegment(pairs [][2]string, segmentationModel string) error {
	if len(pairs) == 0 {
		return nil
	}
	args := []string{
		"kraken",
		"--alto",
	}
	args = append(args, krakenDeviceArgs()...)
	for _, pair := range pairs {
		args = append(args, "-i", pair[0], pair[1])
	}
	args = append(args, "segment", "--yolo", segmentationModel)
	if err := envexec.PythonCmd("yaltai", args...); err != nil {
		return fmt.Errorf("kraken segment failed: %w", err)
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
