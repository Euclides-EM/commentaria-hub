package app

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func applyManualOverride(cfg config, out io.Writer) error {
	if cfg.kind == kindIndex {
		return applyIndexManualOverride(cfg, out)
	}
	return applyLettersManualOverride(cfg, out)
}

func applyIndexManualOverride(cfg config, out io.Writer) error {
	manifest, err := loadIndexManifest(cfg.indexCSV, true)
	if err != nil {
		return err
	}
	pageIndex, err := selectManifestPage(cfg.image, len(manifest.Pages), func(i int) string { return manifest.Pages[i].ImagePath })
	if err != nil {
		return err
	}
	page := &manifest.Pages[pageIndex]
	if cfg.entryNumber > len(page.Entries) {
		return fmt.Errorf("entry %d does not exist in %s (has %d entries)", cfg.entryNumber, page.ImagePath, len(page.Entries))
	}
	entry := &page.Entries[cfg.entryNumber-1]
	changes := map[string]manualFieldChange{}
	changeString(changes, "name", entry.Name, cfg.name, cfg.setFlags["name"], func(v string) { entry.Name = v })
	changeString(changes, "page_number", entry.PageNumber, cfg.pageNumber, cfg.setFlags["page-number"], func(v string) { entry.PageNumber = v })
	changeString(changes, "reference", entry.Reference, cfg.reference, cfg.setFlags["reference"], func(v string) { entry.Reference = v })
	if cfg.setFlags["is-bold"] {
		value, parseErr := strconv.ParseBool(strings.TrimSpace(cfg.isBold))
		if parseErr != nil {
			return fmt.Errorf("--is-bold must be true or false")
		}
		changeString(changes, "is_bold", strconv.FormatBool(entry.IsBold), strconv.FormatBool(value), true, func(string) { entry.IsBold = value })
	}
	if err := validateCorrectedIndexEntry(*entry); err != nil {
		return err
	}
	if len(changes) == 0 {
		return fmt.Errorf("manual override does not change entry %d in %s", cfg.entryNumber, page.ImagePath)
	}
	entry.ManualOverrides = append(entry.ManualOverrides, newManualOverride(cfg, changes))
	if err := saveJSONAtomically(manifestPath(cfg.indexCSV), manifest); err != nil {
		return err
	}
	fmt.Fprintf(out, "Corrected index entry %d in %s (%d fields); run build-csv --kind index to refresh the CSV\n", cfg.entryNumber, page.ImagePath, len(changes))
	return nil
}

func applyLettersManualOverride(cfg config, out io.Writer) error {
	manifest, err := loadLettersManifest(cfg.lettersCSV, true)
	if err != nil {
		return err
	}
	pageIndex, err := selectManifestPage(cfg.image, len(manifest.Pages), func(i int) string { return manifest.Pages[i].ImagePath })
	if err != nil {
		return err
	}
	page := &manifest.Pages[pageIndex]
	if cfg.entryNumber > len(page.Entries) {
		return fmt.Errorf("entry %d does not exist in %s (has %d entries)", cfg.entryNumber, page.ImagePath, len(page.Entries))
	}
	entry := &page.Entries[cfg.entryNumber-1]
	changes := map[string]manualFieldChange{}
	changeString(changes, "letter_number", entry.LetterNumber, cfg.letterNumber, cfg.setFlags["letter-number"], func(v string) { entry.LetterNumber = v })
	changeString(changes, "letter_name", entry.LetterName, cfg.letterName, cfg.setFlags["letter-name"], func(v string) { entry.LetterName = v })
	changeString(changes, "page_number", entry.PageNumber, cfg.pageNumber, cfg.setFlags["page-number"], func(v string) { entry.PageNumber = v })
	if strings.TrimSpace(entry.LetterNumber) == "" || strings.TrimSpace(entry.LetterName) == "" || strings.TrimSpace(entry.PageNumber) == "" {
		return fmt.Errorf("corrected letters entry must have letter_number, letter_name, and page_number")
	}
	if len(changes) == 0 {
		return fmt.Errorf("manual override does not change entry %d in %s", cfg.entryNumber, page.ImagePath)
	}
	entry.ManualOverrides = append(entry.ManualOverrides, newManualOverride(cfg, changes))
	if err := saveJSONAtomically(manifestPath(cfg.lettersCSV), manifest); err != nil {
		return err
	}
	fmt.Fprintf(out, "Corrected letters entry %d in %s (%d fields); run build-csv --kind letters to refresh the CSV\n", cfg.entryNumber, page.ImagePath, len(changes))
	return nil
}

func selectManifestPage(selector string, count int, pathAt func(int) string) (int, error) {
	request := cleanPath(strings.TrimSpace(selector))
	matches := []int{}
	for i := 0; i < count; i++ {
		path := cleanPath(pathAt(i))
		if path == request || filepath.Base(path) == request || strings.HasSuffix(filepath.ToSlash(path), "/"+filepath.ToSlash(request)) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 {
		return 0, fmt.Errorf("image %q was not found in the manifest", selector)
	}
	if len(matches) > 1 {
		return 0, fmt.Errorf("image selector %q is ambiguous; use a longer path", selector)
	}
	return matches[0], nil
}

func changeString(changes map[string]manualFieldChange, field, old, value string, set bool, apply func(string)) {
	if !set {
		return
	}
	value = strings.TrimSpace(value)
	if old == value {
		return
	}
	changes[field] = manualFieldChange{Old: old, New: value}
	apply(value)
}

func validateCorrectedIndexEntry(entry indexEntry) error {
	hasPage := strings.TrimSpace(entry.PageNumber) != ""
	hasReference := strings.TrimSpace(entry.Reference) != ""
	if strings.TrimSpace(entry.Name) == "" || hasPage == hasReference || (hasReference && entry.IsBold) {
		return fmt.Errorf("corrected index entry requires a name and exactly one of page_number or reference; a reference cannot be bold")
	}
	return nil
}

func newManualOverride(cfg config, changes map[string]manualFieldChange) manualOverride {
	return manualOverride{CorrectedBy: strings.TrimSpace(cfg.correctedBy), CorrectedAt: time.Now().UTC(), Reason: strings.TrimSpace(cfg.correctionReason), Changes: changes}
}
