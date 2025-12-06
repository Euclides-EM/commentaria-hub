package krakenwrapper

import (
	"fmt"
	ocrxml "github.com/MiaMish/elements-dh/ocrflow/pkg/xml"
	"os"
	"path/filepath"
)

// FixAltoFileName reads an ALTO XML file from srcPath, modifies the fileName element to contain only the base name of the file, and writes the modified XML to dstPath.
func FixAltoFileName(srcPath, dstPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("could not read ALTO file %s: %w", srcPath, err)
	}
	fixedData := ocrxml.ModifyTag(data, "fileName", func(v string) string {
		return filepath.Base(v)
	})
	if err := os.WriteFile(dstPath, fixedData, 0o644); err != nil {
		return fmt.Errorf("could not write fixed ALTO file to %s: %w", dstPath, err)
	}
	return nil
}
