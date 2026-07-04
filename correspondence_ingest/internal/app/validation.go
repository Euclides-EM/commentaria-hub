package app

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func validateOutputs(cfg config, out io.Writer) error {
	if includesKind(cfg.kind, kindIndex) {
		manifest, err := loadIndexManifest(cfg.indexCSV, true)
		if err != nil {
			return err
		}
		rows := countIndexEntries(manifest)
		if err := validateIndexManifest(manifest, rows); err != nil {
			return err
		}
		fmt.Fprintf(out, "Valid index manifest: %d rows from %d images in %s\n", rows, len(manifest.Pages), manifestPath(cfg.indexCSV))
		reportIssues("index", indexManifestIssues(manifest), out)
		reportFailures("index", indexFailures(manifest), out)
	}
	if includesKind(cfg.kind, kindLetters) {
		manifest, err := loadLettersManifest(cfg.lettersCSV, true)
		if err != nil {
			return err
		}
		rows := countLettersEntries(manifest)
		if err := validateLettersManifest(manifest, rows); err != nil {
			return err
		}
		images, err := discoverImages(cfg.lettersDir)
		if err != nil {
			return fmt.Errorf("validate completed letters-table volumes: %w", err)
		}
		sequenceErr := validateCompletedLetterVolumes(manifest, images)
		fmt.Fprintf(out, "\nValid letters-table manifest: %d rows from %d images in %s\n", rows, len(manifest.Pages), manifestPath(cfg.lettersCSV))
		reportIssues("letters", lettersManifestIssues(manifest), out)
		reportFailures("letters", lettersFailures(manifest), out)
		if sequenceErr != nil {
			return sequenceErr
		}
	}
	return nil
}

func countIndexEntries(manifest indexManifest) int {
	rows := 0
	for _, page := range manifest.Pages {
		rows += len(page.Entries)
	}
	return rows
}

func countLettersEntries(manifest lettersManifest) int {
	rows := 0
	for _, page := range manifest.Pages {
		rows += len(page.Entries)
	}
	return rows
}

func validateCSVFile(path string, expectedHeader []string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return 0, err
	}
	if len(records) == 0 || strings.Join(records[0], "\x00") != strings.Join(expectedHeader, "\x00") {
		return 0, fmt.Errorf("unexpected header in %s; want %s", path, strings.Join(expectedHeader, ","))
	}
	for i, row := range records[1:] {
		if len(row) != len(expectedHeader) {
			return 0, fmt.Errorf("%s row %d has %d columns; want %d", path, i+2, len(row), len(expectedHeader))
		}
		for column, value := range row {
			if strings.TrimSpace(value) == "" && !isOptionalCSVColumn(expectedHeader, expectedHeader[column]) {
				return 0, fmt.Errorf("%s row %d column %s is empty", path, i+2, expectedHeader[column])
			}
		}
		if isIndexCSVHeader(expectedHeader) {
			hasPageNumber := strings.TrimSpace(row[1]) != ""
			hasReference := strings.TrimSpace(row[2]) != ""
			if hasPageNumber == hasReference {
				return 0, fmt.Errorf("%s row %d requires exactly one of page_number or reference", path, i+2)
			}
			if hasReference && row[3] != "false" {
				return 0, fmt.Errorf("%s row %d cross-reference must not be bold", path, i+2)
			}
		}
	}
	return len(records) - 1, nil
}

func isOptionalCSVColumn(header []string, column string) bool {
	return isIndexCSVHeader(header) && (column == "page_number" || column == "reference")
}

func isIndexCSVHeader(header []string) bool {
	return strings.Join(header, "\x00") == strings.Join(indexHeader, "\x00")
}

func validateIndexManifest(manifest indexManifest, csvRows int) error {
	rows := 0
	seen := map[string]bool{}
	for _, page := range manifest.Pages {
		if err := validatePageMetadata(page.ImagePath, page.Volume, page.Provider, page.Model, page.ExtractionMode, page.Transcription, page.TranscriptionProvider, page.TranscriptionModel, page.ExtractedAt, page.TranscribedAt, page.Entries != nil, page.Failure, seen); err != nil {
			return fmt.Errorf("validate index manifest: %w", err)
		}
		for i, entry := range page.Entries {
			hasPageNumber := strings.TrimSpace(entry.PageNumber) != ""
			hasReference := strings.TrimSpace(entry.Reference) != ""
			if strings.TrimSpace(entry.Name) == "" || hasPageNumber == hasReference || (hasReference && entry.IsBold) {
				return fmt.Errorf("validate index manifest: %s entry %d is incomplete", page.ImagePath, i+1)
			}
			if err := validateManualOverrides(entry.ManualOverrides); err != nil {
				return fmt.Errorf("validate index manifest: %s entry %d: %w", page.ImagePath, i+1, err)
			}
		}
		rows += len(page.Entries)
	}
	if rows != csvRows {
		return fmt.Errorf("validate index dataset: manifest has %d entries but CSV has %d rows", rows, csvRows)
	}
	return nil
}

func validateLettersManifest(manifest lettersManifest, csvRows int) error {
	rows := 0
	seen := map[string]bool{}
	for _, page := range manifest.Pages {
		if err := validatePageMetadata(page.ImagePath, page.Volume, page.Provider, page.Model, page.ExtractionMode, page.Transcription, page.TranscriptionProvider, page.TranscriptionModel, page.ExtractedAt, page.TranscribedAt, page.Entries != nil, page.Failure, seen); err != nil {
			return fmt.Errorf("validate letters-table manifest: %w", err)
		}
		for i, entry := range page.Entries {
			if strings.TrimSpace(entry.LetterNumber) == "" || strings.TrimSpace(entry.LetterName) == "" || strings.TrimSpace(entry.PageNumber) == "" {
				return fmt.Errorf("validate letters-table manifest: %s entry %d is incomplete", page.ImagePath, i+1)
			}
			if err := validateManualOverrides(entry.ManualOverrides); err != nil {
				return fmt.Errorf("validate letters-table manifest: %s entry %d: %w", page.ImagePath, i+1, err)
			}
		}
		rows += len(page.Entries)
	}
	if rows != csvRows {
		return fmt.Errorf("validate letters-table dataset: manifest has %d entries but CSV has %d rows", rows, csvRows)
	}
	return nil
}

func validateManualOverrides(overrides []manualOverride) error {
	for i, override := range overrides {
		if strings.TrimSpace(override.CorrectedBy) == "" || override.CorrectedAt.IsZero() || len(override.Changes) == 0 {
			return fmt.Errorf("manual override %d requires corrected_by, corrected_at, and changes", i+1)
		}
		for field, change := range override.Changes {
			if strings.TrimSpace(field) == "" || change.Old == change.New {
				return fmt.Errorf("manual override %d has invalid change for %q", i+1, field)
			}
		}
	}
	return nil
}

type completedLetterVolume struct {
	name    string
	number  int
	minimum int
	maximum int
}

// validateCompletedLetterVolumes applies dataset-level sequence checks only to
// volumes for which every discovered source image has a completed manifest page.
// Supplemental labels such as "836bis" and "Appendice I" are preserved in the
// dataset but do not define or interrupt the ordinary integer letter sequence.
func validateCompletedLetterVolumes(manifest lettersManifest, images []imageInput) error {
	var validationErrors []error
	discovered := map[string]map[string]bool{}
	for _, image := range images {
		if discovered[image.Volume] == nil {
			discovered[image.Volume] = map[string]bool{}
		}
		discovered[image.Volume][cleanPath(image.Path)] = true
	}
	completed := lettersCompleted(manifest)
	entriesByVolume := map[string][]letterEntry{}
	for _, page := range manifest.Pages {
		entriesByVolume[page.Volume] = append(entriesByVolume[page.Volume], page.Entries...)
	}

	volumeNames := make([]string, 0, len(discovered))
	for volume := range discovered {
		volumeNames = append(volumeNames, volume)
	}
	sort.Slice(volumeNames, func(i, j int) bool {
		left, leftOK := parseVolumeNumber(volumeNames[i])
		right, rightOK := parseVolumeNumber(volumeNames[j])
		if leftOK && rightOK && left != right {
			return left < right
		}
		return volumeNames[i] < volumeNames[j]
	})

	var volumes []completedLetterVolume
	for _, volume := range volumeNames {
		paths := discovered[volume]
		complete := true
		for path := range paths {
			if !completed[path] {
				complete = false
				break
			}
		}
		if !complete {
			continue
		}

		volumeNumber, ok := parseVolumeNumber(volume)
		if !ok {
			validationErrors = append(validationErrors, fmt.Errorf("completed volume %q does not have a numeric vol_N name", volume))
			continue
		}
		seen := map[int]letterEntry{}
		pages := map[int]int{}
		minimum, maximum := 0, 0
		for _, entry := range entriesByVolume[volume] {
			number, ordinary := parseOrdinaryLetterNumber(entry.LetterNumber)
			if !ordinary {
				continue
			}
			page, err := strconv.Atoi(strings.TrimSpace(entry.PageNumber))
			if err != nil {
				validationErrors = append(validationErrors, fmt.Errorf("volume %s letter %q has non-numeric page number %q", volume, entry.LetterNumber, entry.PageNumber))
			}
			if previous, duplicate := seen[number]; duplicate {
				validationErrors = append(validationErrors, fmt.Errorf("volume %s has duplicate letter number %d (%q and %q)", volume, number, previous.LetterName, entry.LetterName))
				continue
			}
			seen[number] = entry
			if err == nil {
				pages[number] = page
			}
			if len(seen) == 1 || number < minimum {
				minimum = number
			}
			if len(seen) == 1 || number > maximum {
				maximum = number
			}
		}
		if len(seen) == 0 {
			validationErrors = append(validationErrors, fmt.Errorf("completed volume %s has no ordinary numeric letter numbers", volume))
			continue
		}
		for number := minimum; number <= maximum; number++ {
			if _, ok := seen[number]; !ok {
				validationErrors = append(validationErrors, fmt.Errorf("completed volume %s is missing letter number %d (range %d-%d)", volume, number, minimum, maximum))
			}
		}
		numbers := make([]int, 0, len(pages))
		for number := range pages {
			numbers = append(numbers, number)
		}
		sort.Ints(numbers)
		for i := 1; i < len(numbers); i++ {
			previous, current := numbers[i-1], numbers[i]
			if pages[current] < pages[previous] {
				validationErrors = append(validationErrors, fmt.Errorf("volume %s page numbers regress from letter %d page %d to letter %d page %d", volume, previous, pages[previous], current, pages[current]))
			}
		}
		volumes = append(volumes, completedLetterVolume{name: volume, number: volumeNumber, minimum: minimum, maximum: maximum})
	}

	sort.Slice(volumes, func(i, j int) bool { return volumes[i].number < volumes[j].number })
	for i := 1; i < len(volumes); i++ {
		previous, current := volumes[i-1], volumes[i]
		if current.number != previous.number+1 {
			continue
		}
		if current.minimum != previous.maximum+1 {
			validationErrors = append(validationErrors, fmt.Errorf("completed successive volumes %s and %s are not consecutive: %s ends at %d, %s starts at %d", previous.name, current.name, previous.name, previous.maximum, current.name, current.minimum))
		}
	}
	if len(validationErrors) > 0 {
		messages := make([]string, len(validationErrors))
		for i, err := range validationErrors {
			messages[i] = err.Error()
		}
		return fmt.Errorf("validate letters-table dataset: %d problems\n  - %s", len(messages), strings.Join(messages, "\n  - "))
	}
	return nil
}

func parseOrdinaryLetterNumber(value string) (int, bool) {
	value = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(value), "."))
	number, err := strconv.Atoi(value)
	return number, err == nil
}

func parseVolumeNumber(value string) (int, bool) {
	if !strings.HasPrefix(value, "vol_") {
		return 0, false
	}
	number, err := strconv.Atoi(strings.TrimPrefix(value, "vol_"))
	return number, err == nil
}

func validateIndexCSVMatchesManifest(path string, manifest indexManifest) error {
	expected := [][]string{indexHeader}
	for _, page := range sortIndexPages(manifest.Pages) {
		for _, entry := range page.Entries {
			expected = append(expected, []string{entry.Name, entry.PageNumber, entry.Reference, strconv.FormatBool(entry.IsBold), page.Volume})
		}
	}
	return compareCSVRecords(path, expected)
}

func validateLettersCSVMatchesManifest(path string, manifest lettersManifest) error {
	expected := [][]string{lettersHeader}
	for _, page := range sortLettersPages(manifest.Pages) {
		for _, entry := range page.Entries {
			expected = append(expected, []string{entry.LetterNumber, entry.LetterName, entry.PageNumber, page.Volume})
		}
	}
	return compareCSVRecords(path, expected)
}

func compareCSVRecords(path string, expected [][]string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	actual, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return err
	}
	if len(actual) != len(expected) {
		return fmt.Errorf("validate dataset: %s has %d records; manifest renders %d", path, len(actual), len(expected))
	}
	for row := range expected {
		if len(actual[row]) != len(expected[row]) {
			return fmt.Errorf("validate dataset: %s row %d differs from manifest", path, row+1)
		}
		for column := range expected[row] {
			if actual[row][column] != expected[row][column] {
				return fmt.Errorf("validate dataset: %s row %d column %d differs from manifest", path, row+1, column+1)
			}
		}
	}
	return nil
}

func validatePageMetadata(imagePath, volume, provider, model, extractionMode, transcription, transcriptionProvider, transcriptionModel string, extractedAt, transcribedAt time.Time, complete bool, failure *pageFailure, seen map[string]bool) error {
	path := cleanPath(imagePath)
	if strings.TrimSpace(imagePath) == "" || strings.TrimSpace(volume) == "" {
		return fmt.Errorf("page %q has incomplete provenance", imagePath)
	}
	if extractionMode != modeOnePass && extractionMode != modeTwoPass {
		return fmt.Errorf("page %q has invalid extraction mode %q", imagePath, extractionMode)
	}
	if extractionMode == modeTwoPass && strings.TrimSpace(transcription) != "" && (strings.TrimSpace(transcriptionProvider) == "" || strings.TrimSpace(transcriptionModel) == "" || transcribedAt.IsZero()) {
		return fmt.Errorf("page %q has incomplete transcription provenance", imagePath)
	}
	if complete && failure != nil {
		return fmt.Errorf("page %q is both complete and failed", imagePath)
	}
	if !complete && failure == nil && extractionMode != modeTwoPass {
		return fmt.Errorf("page %q is incomplete without a reusable transcription", imagePath)
	}
	if !complete && failure == nil && strings.TrimSpace(transcription) == "" {
		return fmt.Errorf("page %q is incomplete without a failure or reusable transcription", imagePath)
	}
	if failure != nil {
		if strings.TrimSpace(failure.Error) == "" || strings.TrimSpace(failure.Provider) == "" || strings.TrimSpace(failure.Model) == "" || failure.FailedAt.IsZero() {
			return fmt.Errorf("page %q has incomplete failure provenance", imagePath)
		}
		validPhase := failure.Phase == failurePhaseSingle || failure.Phase == failurePhaseFirst || failure.Phase == failurePhaseSecond
		if !validPhase || (extractionMode == modeOnePass && failure.Phase != failurePhaseSingle) || (failure.Phase == failurePhaseSecond && strings.TrimSpace(transcription) == "") {
			return fmt.Errorf("page %q has invalid failure phase %q", imagePath, failure.Phase)
		}
	}
	if complete && (strings.TrimSpace(provider) == "" || strings.TrimSpace(model) == "" || extractedAt.IsZero()) {
		return fmt.Errorf("page %q has incomplete extraction provenance", imagePath)
	}
	if seen[path] {
		return fmt.Errorf("page %s appears more than once", imagePath)
	}
	seen[path] = true
	return nil
}
