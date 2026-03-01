package krakenwrapper

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
	"golang.org/x/sync/errgroup"
)

// maxDetectLinesWorkers limits concurrent kraken segment processes (CPU-bound subprocess).
func maxDetectLinesWorkers() int {
	n := runtime.NumCPU()
	if n < 2 {
		return 2
	}
	if n > 8 {
		return 8
	}
	return n
}

func DetectLines(imgDir, altoDir string, detectInCategories, ignoreCategories []string) error {
	des, err := os.ReadDir(altoDir)
	if err != nil {
		return err
	}

	var jobs []struct{ imgPath, altoPath string }
	for _, de := range des {
		if !de.IsDir() && filepath.Ext(de.Name()) == ".xml" {
			imgPath := path.Join(imgDir, strings.TrimSuffix(de.Name(), ".xml")+".png")
			altoPath := path.Join(altoDir, de.Name())
			jobs = append(jobs, struct{ imgPath, altoPath string }{imgPath, altoPath})
		}
	}

	if len(jobs) == 0 {
		return nil
	}

	workers := maxDetectLinesWorkers()
	sem := make(chan struct{}, workers)
	grp, ctx := errgroup.WithContext(context.Background())

	for _, job := range jobs {
		job := job
		grp.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sem <- struct{}{}:
				defer func() { <-sem }()
			}
			return detectLinesInFile(job.imgPath, job.altoPath, detectInCategories, ignoreCategories)
		})
	}

	return grp.Wait()
}

func detectLinesInFile(imgPath string, altoPath string, detectInCategories, ignoreCategories []string) error {
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		return fmt.Errorf("the image for ALTO file does not exist in path %s", imgPath)
	}

	// delete existing lines in the ALTO file once (we will glue incrementally)
	if err := alto.DeleteLines(altoPath, altoPath); err != nil {
		return fmt.Errorf("delete lines from ALTO %s: %w", altoPath, err)
	}

	// If no categories were provided, keep the old behavior (single segmentation run).
	if len(detectInCategories) == 0 {
		return detectAndGlueOnce(imgPath, altoPath, detectInCategories, ignoreCategories)
	}

	// Run segmentation per category, gluing lines after each run.
	for _, cat := range detectInCategories {
		// effectiveIgnore = ignoreCategories + (detectInCategories - cat)
		effectiveIgnore := make([]string, 0, len(ignoreCategories)+len(detectInCategories)-1)
		effectiveIgnore = append(effectiveIgnore, ignoreCategories...)
		for _, other := range detectInCategories {
			if other == cat {
				continue
			}
			effectiveIgnore = append(effectiveIgnore, other)
		}
		effectiveIgnore = uniqueNonEmpty(effectiveIgnore)

		// create mask for this category
		maskFile, err := os.CreateTemp("", "mask-*.png")
		if err != nil {
			return err
		}
		maskFile.Close()
		defer os.Remove(maskFile.Name())

		if err := CreateMaskFromALTO(
			altoPath,
			maskFile.Name(),
			[]string{cat},   // detect ONLY this category
			effectiveIgnore, // ignore the rest of detect categories + explicit ignores
		); err != nil {
			return fmt.Errorf("create mask from ALTO %s for category %q: %w", altoPath, cat, err)
		}

		baselineJsonFile, err := os.CreateTemp("", "segmentation-*.json")
		if err != nil {
			return err
		}
		baselineJsonFile.Close()
		defer os.Remove(baselineJsonFile.Name())

		if err := envexec.PythonCmd("kraken",
			"-i", imgPath,
			baselineJsonFile.Name(),
			"segment",
			"-bl",
			"--mask", maskFile.Name(),
			"--pad", "2", "2"); err != nil {
			return fmt.Errorf("kraken segmentation failed for image %s (category %q): %w", imgPath, cat, err)
		}

		if err := GlueLinesToAlto(altoPath, baselineJsonFile.Name(), altoPath); err != nil {
			return fmt.Errorf("glue lines to ALTO %s (category %q): %w", altoPath, cat, err)
		}
	}

	return nil
}

func detectAndGlueOnce(imgPath, altoPath string, detectInCategories, ignoreCategories []string) error {
	// create the mask from the ALTO file
	maskFile, err := os.CreateTemp("", "mask-*.png")
	if err != nil {
		return err
	}
	maskFile.Close()
	defer os.Remove(maskFile.Name())

	if err := CreateMaskFromALTO(altoPath, maskFile.Name(), detectInCategories, ignoreCategories); err != nil {
		return fmt.Errorf("create mask from ALTO %s: %w", altoPath, err)
	}

	baselineJsonFile, err := os.CreateTemp("", "segmentation-*.json")
	if err != nil {
		return err
	}
	baselineJsonFile.Close()
	defer os.Remove(baselineJsonFile.Name())

	if err := envexec.PythonCmd("kraken",
		"-i", imgPath,
		baselineJsonFile.Name(),
		"segment",
		"-bl",
		"--mask", maskFile.Name(),
		"--pad", "2", "2"); err != nil {
		return fmt.Errorf("kraken segmentation failed for image %s: %w", imgPath, err)
	}

	if err := GlueLinesToAlto(altoPath, baselineJsonFile.Name(), altoPath); err != nil {
		return fmt.Errorf("glue lines to ALTO %s: %w", altoPath, err)
	}
	return nil
}

func uniqueNonEmpty(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
