package krakenwrapper

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
)

// RecognizeTextWithMapping overwrites existing ALTO files with OCR-ed ALTO output.
// mapping: key is image filename (relative to inputDir, or just basename), value is absolute or relative ALTO path.
func RecognizeTextWithMapping(imgAndAltoPaths [][2]string, ocrModel string) (<-chan error, error) {
	if strings.TrimSpace(ocrModel) == "" {
		return nil, fmt.Errorf("ocr model is required")
	}
	if err := validateImgAndAltoPaths(imgAndAltoPaths); err != nil {
		return nil, err
	}
	if len(imgAndAltoPaths) == 0 {
		ch := make(chan error, 1)
		close(ch)
		return ch, nil
	}
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- runPairsOCRUsingExistingAlto(imgAndAltoPaths, ocrModel)
	}()
	return errCh, nil
}

func validateImgAndAltoPaths(imgAndAltoPaths [][2]string) error {
	for _, imgAndAltoPath := range imgAndAltoPaths {
		imgPath := imgAndAltoPath[0]
		altoPath := imgAndAltoPath[1]

		// Validate image
		if !slices.Contains(imageFormats, strings.ToLower(filepath.Ext(imgPath))) {
			return fmt.Errorf("image input file %s is not a supported image format (TIFF/PNG)", imgPath)
		}
		if !filepath.IsAbs(imgPath) {
			return fmt.Errorf("image input file %s is not an absolute path", imgPath)
		}
		if _, err := os.Stat(imgPath); err != nil {
			return fmt.Errorf("image input file %s does not exist: %w", imgPath, err)
		}

		// Validate ALTO
		if strings.ToLower(filepath.Ext(altoPath)) != ".xml" {
			return fmt.Errorf("ALTO file %s does not have .xml extension", altoPath)
		}
		if !filepath.IsAbs(altoPath) {
			return fmt.Errorf("ALTO file %s is not an absolute path", altoPath)
		}
		if _, err := os.Stat(altoPath); err != nil {
			return fmt.Errorf("ALTO file %s does not exist: %w", altoPath, err)
		}
	}
	return nil
}

// Uses existing ALTO (segmentation+lines) as input and overwrites it with OCR-ed ALTO.
func runPairsOCRUsingExistingAlto(imgAndAltoPaths [][2]string, ocrModel string) error {

	workers := runtime.NumCPU()
	if workers > maxParallelKraken {
		workers = maxParallelKraken
	}
	if workers > len(imgAndAltoPaths) {
		workers = len(imgAndAltoPaths)
	}
	if workers < 1 {
		workers = 1
	}
	chunks := chunkPairs(imgAndAltoPaths, workers)

	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	wg.Add(len(chunks))
	for _, chunk := range chunks {
		chunk := chunk
		go func() {
			defer wg.Done()
			if err := runKrakenOCRReuseAlto(chunk, ocrModel); err != nil {
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

	// Rename temp outputs into place, then fix fileName (parallel by file).
	var postWg sync.WaitGroup
	var postMu sync.Mutex
	var postErr error

	postWg.Add(len(imgAndAltoPaths))
	for _, pair := range imgAndAltoPaths {
		altoPath := pair[1]
		go func(final string) {
			defer postWg.Done()

			tmp := tmpOcredPath(final)

			if err := RemovePathFromAltoImgFileName(tmp, tmp); err != nil {
				postMu.Lock()
				if postErr == nil {
					postErr = fmt.Errorf("could not fix ALTO file name in %s: %w", final, err)
				}
				postMu.Unlock()
				return
			}

			// Replace final with tmp (best effort atomic on same filesystem).
			_ = os.Remove(final)
			if err := os.Rename(tmp, final); err != nil {
				postMu.Lock()
				if postErr == nil {
					postErr = fmt.Errorf("could not replace ALTO %s with OCR output: %w", final, err)
				}
				postMu.Unlock()
				return
			}
		}(altoPath)
	}
	postWg.Wait()

	return postErr
}

func runKrakenOCRReuseAlto(pairs [][2]string, ocrModel string) error {
	if len(pairs) == 0 {
		return nil
	}

	var args = []string{"--alto"}

	// For each pair: -i <img> <existing_alto_in> <alto_out_tmp>
	for _, p := range pairs {
		imgPath := p[0]
		altoOutTmp := tmpOcredPath(p[1])

		// Ensure output dir exists
		if err := os.MkdirAll(filepath.Dir(altoOutTmp), 0755); err != nil {
			return fmt.Errorf("could not create ALTO output directory %s: %w", filepath.Dir(altoOutTmp), err)
		}

		args = append(args, "-i", imgPath, altoOutTmp)
	}

	args = append(args, "ocr", "-m", ocrModel)

	if err := envexec.PythonCmd("kraken", args...); err != nil {
		return fmt.Errorf("kraken ocr (reuse alto) failed: %w", err)
	}

	return nil
}

func tmpOcredPath(finalAltoPath string) string {
	return finalAltoPath + ".ocr.tmp"
}
