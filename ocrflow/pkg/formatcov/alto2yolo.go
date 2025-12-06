package formatcov

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"os"
	"path/filepath"
	"strings"
)

func Alto2Yolo(imgDir, altoDir, outputDir string, shuffle float64, segmontoGranularity string) error {
	var imagesInAlto []string
	altoFiles, err := os.ReadDir(altoDir)
	if err != nil {
		return fmt.Errorf("failed to read ALTO dir: %w", err)
	}
	for _, file := range altoFiles {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".xml" {
			imagesInAlto = append(imagesInAlto, filepath.Join(imgDir, strings.TrimSuffix(file.Name(), ".xml")+".png"))
		}
	}

	tmpDir, err := os.MkdirTemp("", "alto2yolo-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	if err = futils.CopyDir(altoDir, tmpDir); err != nil {
		return fmt.Errorf("failed to copy ALTO files to temp dir: %w", err)
	}
	for _, image := range imagesInAlto {
		destImagePath := filepath.Join(tmpDir, filepath.Base(image))
		if err = futils.CopyFile(image, destImagePath); err != nil {
			return fmt.Errorf("failed to copy image %s to temp dir: %w", image, err)
		}
	}

	c := "yaltai convert alto-to-yolo " + tmpDir + "/*.xml " + outputDir
	if shuffle > 0 {
		c += " --shuffle " + fmt.Sprintf("%.2f", shuffle)
	}
	if segmontoGranularity != "" {
		c += " --segmonto " + segmontoGranularity
	}

	return envexec.Cmd("bash", "-c", c)
}
