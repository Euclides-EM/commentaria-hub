package formatcov

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	ocrxml "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/xml"
)

func setAltoImageFileName(altoPath, imageName string) error {
	data, err := os.ReadFile(altoPath)
	if err != nil {
		return fmt.Errorf("failed to read ALTO file %s: %w", altoPath, err)
	}
	if _, err := ocrxml.ExtractTagValue(string(data), "fileName"); err != nil {
		return fmt.Errorf("failed to locate source image filename in ALTO file %s: %w", altoPath, err)
	}
	data = ocrxml.ModifyTag(data, "fileName", func(string) string {
		return imageName
	})
	if err := os.WriteFile(altoPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to update source image filename in ALTO file %s: %w", altoPath, err)
	}
	return nil
}

func Alto2Yolo(imgDir, altoDir, outputDir string, shuffle float64, segmontoGranularity string) error {
	var imagesInAlto []string
	altoFiles, err := os.ReadDir(altoDir)
	if err != nil {
		return fmt.Errorf("failed to read ALTO dir: %w", err)
	}
	for _, file := range altoFiles {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".xml" && file.Name() != "METS.xml" {
			imagesInAlto = append(imagesInAlto, filepath.Join(imgDir, strings.TrimSuffix(file.Name(), ".xml")+".png"))
		}
	}

	tmpDir, err := futils.MkdirTemp("alto2yolo")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	if err = futils.CopyDir(altoDir, tmpDir); err != nil {
		return fmt.Errorf("failed to copy ALTO files to temp dir: %w", err)
	}
	// yaltai/kraken only accept ALTO or PAGE XML; METS is a manifest format and must be excluded
	if err := os.Remove(filepath.Join(tmpDir, "METS.xml")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove METS.xml from temp dir: %w", err)
	}
	for _, image := range imagesInAlto {
		destImagePath := filepath.Join(tmpDir, filepath.Base(image))
		if err = futils.CopyFile(image, destImagePath); err != nil {
			return fmt.Errorf("failed to copy image %s to temp dir: %w", image, err)
		}
		altoName := strings.TrimSuffix(filepath.Base(image), filepath.Ext(image)) + ".xml"
		if err = setAltoImageFileName(filepath.Join(tmpDir, altoName), filepath.Base(destImagePath)); err != nil {
			return err
		}
	}

	c := "yaltai convert alto-to-yolo " + tmpDir + "/*.xml " + outputDir
	if shuffle > 0 {
		c += " --shuffle " + fmt.Sprintf("%.2f", shuffle)
	}
	if segmontoGranularity != "" {
		c += " --segmonto " + segmontoGranularity
	}

	if err := envexec.PythonBashCmd(c); err != nil {
		if cleanupErr := os.RemoveAll(outputDir); cleanupErr != nil {
			return fmt.Errorf("ALTO-to-YOLO conversion failed: %w; failed to remove partial output %s: %v", err, outputDir, cleanupErr)
		}
		return err
	}
	return nil
}
