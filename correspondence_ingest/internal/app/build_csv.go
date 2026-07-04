package app

import (
	"fmt"
	"io"
	"os"
)

// buildCSVOutputs creates disposable CSV exports from the manifests.
func buildCSVOutputs(cfg config, out io.Writer) error {
	if includesKind(cfg.kind, kindIndex) {
		if _, err := os.Stat(manifestPath(cfg.indexCSV)); err != nil {
			return fmt.Errorf("build index CSV: manifest %s: %w", manifestPath(cfg.indexCSV), err)
		}
		manifest, err := loadIndexManifest(cfg.indexCSV, true)
		if err != nil {
			return err
		}
		if err := renderIndexCSV(cfg.indexCSV, manifest); err != nil {
			return fmt.Errorf("build index CSV: %w", err)
		}
		if _, err := validateCSVFile(cfg.indexCSV, indexHeader); err != nil {
			return fmt.Errorf("validate built index CSV: %w", err)
		}
		fmt.Fprintf(out, "Built index CSV from manifest: %s\n", cfg.indexCSV)
	}
	if includesKind(cfg.kind, kindLetters) {
		if _, err := os.Stat(manifestPath(cfg.lettersCSV)); err != nil {
			return fmt.Errorf("build letters-table CSV: manifest %s: %w", manifestPath(cfg.lettersCSV), err)
		}
		manifest, err := loadLettersManifest(cfg.lettersCSV, true)
		if err != nil {
			return err
		}
		if err := renderLettersCSV(cfg.lettersCSV, manifest); err != nil {
			return fmt.Errorf("build letters-table CSV: %w", err)
		}
		if _, err := validateCSVFile(cfg.lettersCSV, lettersHeader); err != nil {
			return fmt.Errorf("validate built letters-table CSV: %w", err)
		}
		fmt.Fprintf(out, "Built letters-table CSV from manifest: %s\n", cfg.lettersCSV)
	}
	return nil
}
