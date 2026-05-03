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
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
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

	// create temp dir that includes the ato and images
	tmpDir, err := futils.MkdirTemp("kraken_ocr_reuse_alto")
	if err != nil {
		return fmt.Errorf("could not create temp dir for kraken OCR reuse ALTO: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	tmpToOrigAltoPath := make(map[string]string) // tmp ALTO path -> original ALTO path
	imgAndAltoTmpPaths := make([][2]string, 0)
	for _, pair := range imgAndAltoPaths {
		imgPath := pair[0]
		altoPath := pair[1]

		tmpImgPath := filepath.Join(tmpDir, filepath.Base(imgPath))
		tmpAltoPath := filepath.Join(tmpDir, filepath.Base(altoPath))

		if err := futils.CopyFile(imgPath, tmpImgPath); err != nil {
			return fmt.Errorf("could not link image %s to temp location: %w", imgPath, err)
		}
		if err := futils.CopyFile(altoPath, tmpAltoPath); err != nil {
			return fmt.Errorf("could not link ALTO %s to temp location: %w", altoPath, err)
		}

		imgAndAltoTmpPaths = append(imgAndAltoTmpPaths, [2]string{tmpImgPath, tmpAltoPath})
		tmpToOrigAltoPath[tmpAltoPath] = altoPath
	}

	// Run Kraken OCR in parallel on pairs of (img, alto) files. Kraken will read the ALTO for segmentation

	workers := runtime.NumCPU()
	if workers > maxParallelKraken {
		workers = maxParallelKraken
	}
	if workers > len(imgAndAltoTmpPaths) {
		workers = len(imgAndAltoTmpPaths)
	}
	if workers < 1 {
		workers = 1
	}
	chunks := chunkPairs(imgAndAltoTmpPaths, workers)

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

	postWg.Add(len(imgAndAltoTmpPaths))
	for _, pair := range imgAndAltoTmpPaths {
		altoTmpPath := pair[1]
		go func(altoTmpPath string) {
			defer postWg.Done()

			tmp := tmpOcredPath(altoTmpPath)
			if _, err := os.Stat(tmp); err != nil {
				postMu.Lock()
				if postErr == nil {
					postErr = fmt.Errorf("expected OCR output file %s does not exist: %w", tmp, err)
				}
				postMu.Unlock()
				return
			}

			// Replace final with tmp (best effort atomic on same filesystem).
			_ = os.Remove(tmpToOrigAltoPath[altoTmpPath])
			if err := os.Rename(tmp, tmpToOrigAltoPath[altoTmpPath]); err != nil {
				postMu.Lock()
				if postErr == nil {
					postErr = fmt.Errorf("could not replace ALTO %s with OCR output: %w", altoTmpPath, err)
				}
				postMu.Unlock()
				return
			}
		}(altoTmpPath)
	}
	postWg.Wait()

	return postErr
}

func runKrakenOCRReuseAlto(pairs [][2]string, ocrModel string) error {
	// todo - this does not work!! Kraken CLI doesn't have the option to load existing ALTO for reuse.
	//  The way to resolve it is to use a Python script - I started working on it but it is not ready yet.
	//  See it in python-tools/ocr_for_segmented.py
	//  Probably, much of my work to have absolute paths and temp files can be removed once we have the script...

	if len(pairs) == 0 {
		return nil
	}

	var args = []string{"--alto"}
	args = append(args, krakenDeviceArgs()...)

	// For each pair: -i <img> <alto_out_tmp>
	for _, p := range pairs {
		imgPath := p[0]
		altoOutTmp := tmpOcredPath(p[1])

		// Ensure output dir exists
		if err := os.MkdirAll(filepath.Dir(altoOutTmp), 0755); err != nil {
			return fmt.Errorf("could not create ALTO output directory %s: %w", filepath.Dir(altoOutTmp), err)
		}

		args = append(args, "-i", imgPath, altoOutTmp)
	}

	args = append(args, "segment", "-t", "alto", "ocr", "-m", ocrModel)

	if err := envexec.PythonCmd("kraken", args...); err != nil {
		return fmt.Errorf("kraken ocr (reuse alto) failed: %w", err)
	}

	return nil
}

func tmpOcredPath(finalAltoPath string) string {
	return finalAltoPath + ".ocr.tmp"
}
