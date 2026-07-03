package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const manifestVersion = 1

var (
	indexHeader   = []string{"name", "page_number", "is_bold", "volume"}
	lettersHeader = []string{"letter_number", "letter_name", "page_number", "volume"}
)

type indexManifest struct {
	Version int               `json:"version"`
	Kind    string            `json:"kind"`
	Pages   []indexPageResult `json:"pages"`
}

type indexPageResult struct {
	ImagePath   string       `json:"image_path"`
	Volume      string       `json:"volume"`
	Provider    string       `json:"provider"`
	Model       string       `json:"model"`
	ExtractedAt time.Time    `json:"extracted_at"`
	Entries     []indexEntry `json:"entries"`
}

type lettersManifest struct {
	Version int                 `json:"version"`
	Kind    string              `json:"kind"`
	Pages   []lettersPageResult `json:"pages"`
}

type lettersPageResult struct {
	ImagePath   string        `json:"image_path"`
	Volume      string        `json:"volume"`
	Provider    string        `json:"provider"`
	Model       string        `json:"model"`
	ExtractedAt time.Time     `json:"extracted_at"`
	Entries     []letterEntry `json:"entries"`
}

func runIndexExtraction(cfg config, client llmExecutor, out io.Writer) error {
	images, err := discoverImages(cfg.indexDir)
	if err != nil {
		return fmt.Errorf("discover index images: %w", err)
	}
	selected, targeted, err := selectImages(images, cfg.rerunImages)
	if err != nil {
		return fmt.Errorf("select index images: %w", err)
	}
	manifest, err := loadIndexManifest(cfg.indexCSV, cfg.resume)
	if err != nil {
		return err
	}
	if !cfg.resume {
		if err := saveJSONAtomically(manifestPath(cfg.indexCSV), manifest); err != nil {
			return fmt.Errorf("initialize index manifest: %w", err)
		}
		if err := renderIndexCSV(cfg.indexCSV, manifest); err != nil {
			return fmt.Errorf("initialize index CSV: %w", err)
		}
	}
	completed := indexCompleted(manifest)
	written, skipped := 0, 0
	for i, image := range selected {
		if cfg.resume && !targeted && completed[cleanPath(image.Path)] {
			skipped++
			fmt.Fprintf(out, "[%d/%d] Skipping completed index page %s\n", i+1, len(selected), image.Path)
			continue
		}
		fmt.Fprintf(out, "[%d/%d] Extracting index page %s\n", i+1, len(selected), image.Path)
		raw, err := client.Exec(cfg.provider, cfg.model, indexPrompt, image.Path)
		if err != nil {
			return fmt.Errorf("extract index image %s: %w", image.Path, err)
		}
		entries, err := parseIndexResponse(raw, image.Volume)
		if err != nil {
			return fmt.Errorf("parse index response for %s: %w", image.Path, err)
		}
		manifest.Pages = upsertIndexPage(manifest.Pages, indexPageResult{
			ImagePath: cleanPath(image.Path), Volume: image.Volume, Provider: cfg.provider,
			Model: cfg.model, ExtractedAt: time.Now().UTC(), Entries: entries,
		})
		if err := saveJSONAtomically(manifestPath(cfg.indexCSV), manifest); err != nil {
			return fmt.Errorf("save index manifest after %s: %w", image.Path, err)
		}
		if err := renderIndexCSV(cfg.indexCSV, manifest); err != nil {
			return fmt.Errorf("render index CSV after %s: %w", image.Path, err)
		}
		written += len(entries)
	}
	if err := renderIndexCSV(cfg.indexCSV, manifest); err != nil {
		return fmt.Errorf("render index CSV: %w", err)
	}
	fmt.Fprintf(out, "Index extraction complete: %d rows extracted, %d images skipped; manifest %s; output %s\n", written, skipped, manifestPath(cfg.indexCSV), cfg.indexCSV)
	return nil
}

func runLettersExtraction(cfg config, client llmExecutor, out io.Writer) error {
	images, err := discoverImages(cfg.lettersDir)
	if err != nil {
		return fmt.Errorf("discover letters-table images: %w", err)
	}
	selected, targeted, err := selectImages(images, cfg.rerunImages)
	if err != nil {
		return fmt.Errorf("select letters-table images: %w", err)
	}
	manifest, err := loadLettersManifest(cfg.lettersCSV, cfg.resume)
	if err != nil {
		return err
	}
	if !cfg.resume {
		if err := saveJSONAtomically(manifestPath(cfg.lettersCSV), manifest); err != nil {
			return fmt.Errorf("initialize letters-table manifest: %w", err)
		}
		if err := renderLettersCSV(cfg.lettersCSV, manifest); err != nil {
			return fmt.Errorf("initialize letters-table CSV: %w", err)
		}
	}
	completed := lettersCompleted(manifest)
	written, skipped := 0, 0
	for i, image := range selected {
		if cfg.resume && !targeted && completed[cleanPath(image.Path)] {
			skipped++
			fmt.Fprintf(out, "[%d/%d] Skipping completed letters-table page %s\n", i+1, len(selected), image.Path)
			continue
		}
		fmt.Fprintf(out, "[%d/%d] Extracting letters-table page %s\n", i+1, len(selected), image.Path)
		raw, err := client.Exec(cfg.provider, cfg.model, lettersTablePrompt, image.Path)
		if err != nil {
			return fmt.Errorf("extract letters-table image %s: %w", image.Path, err)
		}
		entries, err := parseLettersResponse(raw, image.Volume)
		if err != nil {
			return fmt.Errorf("parse letters-table response for %s: %w", image.Path, err)
		}
		manifest.Pages = upsertLettersPage(manifest.Pages, lettersPageResult{
			ImagePath: cleanPath(image.Path), Volume: image.Volume, Provider: cfg.provider,
			Model: cfg.model, ExtractedAt: time.Now().UTC(), Entries: entries,
		})
		if err := saveJSONAtomically(manifestPath(cfg.lettersCSV), manifest); err != nil {
			return fmt.Errorf("save letters-table manifest after %s: %w", image.Path, err)
		}
		if err := renderLettersCSV(cfg.lettersCSV, manifest); err != nil {
			return fmt.Errorf("render letters-table CSV after %s: %w", image.Path, err)
		}
		written += len(entries)
	}
	if err := renderLettersCSV(cfg.lettersCSV, manifest); err != nil {
		return fmt.Errorf("render letters-table CSV: %w", err)
	}
	fmt.Fprintf(out, "Letters-table extraction complete: %d rows extracted, %d images skipped; manifest %s; output %s\n", written, skipped, manifestPath(cfg.lettersCSV), cfg.lettersCSV)
	return nil
}

func manifestPath(outputPath string) string { return outputPath + ".manifest.json" }

func loadIndexManifest(outputPath string, resume bool) (indexManifest, error) {
	manifest := indexManifest{Version: manifestVersion, Kind: kindIndex, Pages: []indexPageResult{}}
	if !resume {
		return manifest, nil
	}
	if err := loadJSON(manifestPath(outputPath), &manifest); errors.Is(err, os.ErrNotExist) {
		if hasRows, checkErr := existingCSVHasRows(outputPath); checkErr != nil {
			return manifest, checkErr
		} else if hasRows {
			return manifest, fmt.Errorf("index output %s predates its manifest; use --rerun to create agent-safe state", outputPath)
		}
		return manifest, nil
	} else if err != nil {
		return manifest, fmt.Errorf("load index manifest: %w", err)
	}
	if manifest.Version != manifestVersion || manifest.Kind != kindIndex {
		return manifest, fmt.Errorf("invalid index manifest metadata in %s", manifestPath(outputPath))
	}
	return manifest, nil
}

func loadLettersManifest(outputPath string, resume bool) (lettersManifest, error) {
	manifest := lettersManifest{Version: manifestVersion, Kind: kindLetters, Pages: []lettersPageResult{}}
	if !resume {
		return manifest, nil
	}
	if err := loadJSON(manifestPath(outputPath), &manifest); errors.Is(err, os.ErrNotExist) {
		if hasRows, checkErr := existingCSVHasRows(outputPath); checkErr != nil {
			return manifest, checkErr
		} else if hasRows {
			return manifest, fmt.Errorf("letters-table output %s predates its manifest; use --rerun to create agent-safe state", outputPath)
		}
		return manifest, nil
	} else if err != nil {
		return manifest, fmt.Errorf("load letters-table manifest: %w", err)
	}
	if manifest.Version != manifestVersion || manifest.Kind != kindLetters {
		return manifest, fmt.Errorf("invalid letters-table manifest metadata in %s", manifestPath(outputPath))
	}
	return manifest, nil
}

func loadJSON(path string, destination any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	return nil
}

func saveJSONAtomically(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomically(path, func(file *os.File) error {
		_, err := file.Write(data)
		return err
	})
}

func upsertIndexPage(pages []indexPageResult, page indexPageResult) []indexPageResult {
	for i := range pages {
		if cleanPath(pages[i].ImagePath) == page.ImagePath {
			pages[i] = page
			return sortIndexPages(pages)
		}
	}
	return sortIndexPages(append(pages, page))
}

func upsertLettersPage(pages []lettersPageResult, page lettersPageResult) []lettersPageResult {
	for i := range pages {
		if cleanPath(pages[i].ImagePath) == page.ImagePath {
			pages[i] = page
			return sortLettersPages(pages)
		}
	}
	return sortLettersPages(append(pages, page))
}

func sortIndexPages(pages []indexPageResult) []indexPageResult {
	sort.Slice(pages, func(i, j int) bool { return pages[i].ImagePath < pages[j].ImagePath })
	return pages
}

func sortLettersPages(pages []lettersPageResult) []lettersPageResult {
	sort.Slice(pages, func(i, j int) bool { return pages[i].ImagePath < pages[j].ImagePath })
	return pages
}

func indexCompleted(manifest indexManifest) map[string]bool {
	completed := make(map[string]bool, len(manifest.Pages))
	for _, page := range manifest.Pages {
		completed[cleanPath(page.ImagePath)] = true
	}
	return completed
}

func lettersCompleted(manifest lettersManifest) map[string]bool {
	completed := make(map[string]bool, len(manifest.Pages))
	for _, page := range manifest.Pages {
		completed[cleanPath(page.ImagePath)] = true
	}
	return completed
}

func selectImages(images []imageInput, selectors string) ([]imageInput, bool, error) {
	requested := splitCommaList(selectors)
	if len(requested) == 0 {
		return images, false, nil
	}
	selected := make([]imageInput, 0, len(requested))
	seen := map[string]bool{}
	for _, request := range requested {
		var matches []imageInput
		for _, image := range images {
			path := cleanPath(image.Path)
			requestPath := cleanPath(request)
			if path == requestPath || filepath.Base(path) == requestPath || strings.HasSuffix(filepath.ToSlash(path), "/"+filepath.ToSlash(requestPath)) {
				matches = append(matches, image)
			}
		}
		if len(matches) == 0 {
			return nil, true, fmt.Errorf("image %q was not found", request)
		}
		if len(matches) > 1 {
			return nil, true, fmt.Errorf("image selector %q is ambiguous; use a longer path", request)
		}
		path := cleanPath(matches[0].Path)
		if !seen[path] {
			selected = append(selected, matches[0])
			seen[path] = true
		}
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Path < selected[j].Path })
	return selected, true, nil
}

func splitCommaList(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func cleanPath(path string) string { return filepath.Clean(path) }

func renderIndexCSV(path string, manifest indexManifest) error {
	return writeCSVAtomically(path, indexHeader, func(writer *csv.Writer) error {
		for _, page := range sortIndexPages(manifest.Pages) {
			for _, entry := range page.Entries {
				if err := writer.Write([]string{entry.Name, entry.PageNumber, strconv.FormatBool(entry.IsBold), page.Volume}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func renderLettersCSV(path string, manifest lettersManifest) error {
	return writeCSVAtomically(path, lettersHeader, func(writer *csv.Writer) error {
		for _, page := range sortLettersPages(manifest.Pages) {
			for _, entry := range page.Entries {
				if err := writer.Write([]string{entry.LetterNumber, entry.LetterName, entry.PageNumber, page.Volume}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func writeCSVAtomically(path string, header []string, rows func(*csv.Writer) error) error {
	return writeFileAtomically(path, func(file *os.File) error {
		writer := csv.NewWriter(file)
		if err := writer.Write(header); err != nil {
			return err
		}
		if err := rows(writer); err != nil {
			return err
		}
		writer.Flush()
		return writer.Error()
	})
}

func writeFileAtomically(path string, write func(*os.File) error) (returnErr error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".indexextractor-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() {
		if err := os.Remove(temporaryPath); err != nil && !errors.Is(err, os.ErrNotExist) && returnErr == nil {
			returnErr = err
		}
	}()
	if err := write(temporary); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func existingCSVHasRows(path string) (bool, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil {
		return false, err
	}
	_, err = reader.Read()
	return err == nil, nil
}

func validateOutputs(cfg config, out io.Writer) error {
	if includesKind(cfg.kind, kindIndex) {
		rows, err := validateCSVFile(cfg.indexCSV, indexHeader)
		if err != nil {
			return fmt.Errorf("validate index output: %w", err)
		}
		if _, err := os.Stat(manifestPath(cfg.indexCSV)); errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(out, "Valid legacy index CSV: %d rows in %s (no manifest)\n", rows, cfg.indexCSV)
		} else if err != nil {
			return err
		} else {
			manifest, err := loadIndexManifest(cfg.indexCSV, true)
			if err != nil {
				return err
			}
			if err := validateIndexManifest(manifest, rows); err != nil {
				return err
			}
			if err := validateIndexCSVMatchesManifest(cfg.indexCSV, manifest); err != nil {
				return err
			}
			fmt.Fprintf(out, "Valid index dataset: %d rows from %d images in %s\n", rows, len(manifest.Pages), cfg.indexCSV)
		}
	}
	if includesKind(cfg.kind, kindLetters) {
		rows, err := validateCSVFile(cfg.lettersCSV, lettersHeader)
		if err != nil {
			return fmt.Errorf("validate letters-table output: %w", err)
		}
		if _, err := os.Stat(manifestPath(cfg.lettersCSV)); errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(out, "Valid legacy letters-table CSV: %d rows in %s (no manifest)\n", rows, cfg.lettersCSV)
		} else if err != nil {
			return err
		} else {
			manifest, err := loadLettersManifest(cfg.lettersCSV, true)
			if err != nil {
				return err
			}
			if err := validateLettersManifest(manifest, rows); err != nil {
				return err
			}
			if err := validateLettersCSVMatchesManifest(cfg.lettersCSV, manifest); err != nil {
				return err
			}
			fmt.Fprintf(out, "Valid letters-table dataset: %d rows from %d images in %s\n", rows, len(manifest.Pages), cfg.lettersCSV)
		}
	}
	return nil
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
			if strings.TrimSpace(value) == "" {
				return 0, fmt.Errorf("%s row %d column %s is empty", path, i+2, expectedHeader[column])
			}
		}
	}
	return len(records) - 1, nil
}

func validateIndexManifest(manifest indexManifest, csvRows int) error {
	rows := 0
	seen := map[string]bool{}
	for _, page := range manifest.Pages {
		if err := validatePageMetadata(page.ImagePath, page.Volume, page.Provider, page.Model, page.ExtractedAt, seen); err != nil {
			return fmt.Errorf("validate index manifest: %w", err)
		}
		for i, entry := range page.Entries {
			if strings.TrimSpace(entry.Name) == "" || strings.TrimSpace(entry.PageNumber) == "" {
				return fmt.Errorf("validate index manifest: %s entry %d is incomplete", page.ImagePath, i+1)
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
		if err := validatePageMetadata(page.ImagePath, page.Volume, page.Provider, page.Model, page.ExtractedAt, seen); err != nil {
			return fmt.Errorf("validate letters-table manifest: %w", err)
		}
		for i, entry := range page.Entries {
			if strings.TrimSpace(entry.LetterNumber) == "" || strings.TrimSpace(entry.LetterName) == "" || strings.TrimSpace(entry.PageNumber) == "" {
				return fmt.Errorf("validate letters-table manifest: %s entry %d is incomplete", page.ImagePath, i+1)
			}
		}
		rows += len(page.Entries)
	}
	if rows != csvRows {
		return fmt.Errorf("validate letters-table dataset: manifest has %d entries but CSV has %d rows", rows, csvRows)
	}
	return nil
}

func validateIndexCSVMatchesManifest(path string, manifest indexManifest) error {
	expected := [][]string{indexHeader}
	for _, page := range sortIndexPages(manifest.Pages) {
		for _, entry := range page.Entries {
			expected = append(expected, []string{entry.Name, entry.PageNumber, strconv.FormatBool(entry.IsBold), page.Volume})
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

func validatePageMetadata(imagePath, volume, provider, model string, extractedAt time.Time, seen map[string]bool) error {
	path := cleanPath(imagePath)
	if strings.TrimSpace(imagePath) == "" || strings.TrimSpace(volume) == "" || strings.TrimSpace(provider) == "" || strings.TrimSpace(model) == "" || extractedAt.IsZero() {
		return fmt.Errorf("page %q has incomplete provenance", imagePath)
	}
	if seen[path] {
		return fmt.Errorf("page %s appears more than once", imagePath)
	}
	seen[path] = true
	return nil
}

func reportStatus(cfg config, out io.Writer) error {
	if includesKind(cfg.kind, kindIndex) {
		images, err := discoverImages(cfg.indexDir)
		if err != nil {
			return fmt.Errorf("status index: %w", err)
		}
		manifest, err := loadIndexManifest(cfg.indexCSV, true)
		if err != nil {
			return err
		}
		reportCompletion("index", images, indexCompleted(manifest), cfg.indexCSV, out)
	}
	if includesKind(cfg.kind, kindLetters) {
		images, err := discoverImages(cfg.lettersDir)
		if err != nil {
			return fmt.Errorf("status letters: %w", err)
		}
		manifest, err := loadLettersManifest(cfg.lettersCSV, true)
		if err != nil {
			return err
		}
		reportCompletion("letters", images, lettersCompleted(manifest), cfg.lettersCSV, out)
	}
	return nil
}

func reportCompletion(label string, images []imageInput, completed map[string]bool, outputPath string, out io.Writer) {
	known := 0
	for _, image := range images {
		if completed[cleanPath(image.Path)] {
			known++
		}
	}
	fmt.Fprintf(out, "%s: %d/%d images completed, %d pending; manifest %s\n", label, known, len(images), len(images)-known, manifestPath(outputPath))
}
