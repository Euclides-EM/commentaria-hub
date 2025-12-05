package krakenwrapper

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// FixAltoFileName reads an ALTO XML file from srcPath, modifies the fileName element to contain only the base name of the file, and writes the modified XML to dstPath.
func FixAltoFileName(srcPath, dstPath string) error {
	// search for the <fileName> element and change its content to the base name of the file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("could not read ALTO file %s: %w", srcPath, err)
	}
	re := regexp.MustCompile(`<fileName>([^<]+)</fileName>`)
	match := re.FindStringSubmatch(string(data))
	if len(match) <= 1 {
		return fmt.Errorf("could not parse ALTO file %s", srcPath)
	}
	filename := match[1]
	fixedFileName := filepath.Base(filename)
	fixedData := re.ReplaceAll(data, []byte(fmt.Sprintf("<fileName>%s</fileName>", fixedFileName)))
	if err := os.WriteFile(dstPath, fixedData, 0o644); err != nil {
		return fmt.Errorf("could not write fixed ALTO file to %s: %w", dstPath, err)
	}
	return nil

}
