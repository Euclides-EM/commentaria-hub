package krakenwrapper

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
)

func DetectLines(imgDir, altoDir string, detectInCategories, ignoreCategories []string) error {
	des, err := os.ReadDir(altoDir)
	if err != nil {
		return err
	}

	for _, de := range des {
		if !de.IsDir() && filepath.Ext(de.Name()) == ".xml" {
			imgPath := path.Join(imgDir, strings.TrimSuffix(de.Name(), ".xml")+".png")
			altoPath := path.Join(altoDir, de.Name())

			err2 := detectLinesInFile(imgPath, altoPath, detectInCategories, ignoreCategories)
			if err2 != nil {
				return err2
			}
		}
	}
	return nil
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
