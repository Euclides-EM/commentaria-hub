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

	// create the mask from the ALTO file
	maskFile, err := os.CreateTemp("", "mask-*.png")
	if err != nil {
		return err
	}
	maskFile.Close()
	defer os.Remove(maskFile.Name())

	if err := CreateMaskFromALTO(altoPath, maskFile.Name(),
		detectInCategories,
		ignoreCategories); err != nil {
		return fmt.Errorf("create mask from ALTO %s: %w", altoPath, err)
	}

	// delete existing lines in the ALTO file
	if err := alto.DeleteLines(altoPath, altoPath); err != nil {
		return fmt.Errorf("delete lines from ALTO %s: %w", altoPath, err)
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
