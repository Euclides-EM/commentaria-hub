package app

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func manifestPath(outputPath string) string { return outputPath + ".manifest.json" }

func loadIndexManifest(outputPath string, resume bool) (indexManifest, error) {
	manifest := indexManifest{Version: manifestVersion, Kind: kindIndex, Pages: []indexPageResult{}}
	if !resume {
		return manifest, nil
	}
	if err := loadJSON(manifestPath(outputPath), &manifest); errors.Is(err, os.ErrNotExist) {
		return manifest, nil
	} else if err != nil {
		return manifest, fmt.Errorf("load index manifest: %w", err)
	}
	if manifest.Kind != kindIndex || (manifest.Version != legacyManifestVersion && manifest.Version != twoPassManifestVersion && manifest.Version != resumableManifestVersion && manifest.Version != failureManifestVersion && manifest.Version != manifestVersion) {
		return manifest, fmt.Errorf("invalid index manifest metadata in %s", manifestPath(outputPath))
	}
	migrateIndexManifest(&manifest)
	return manifest, nil
}

func loadLettersManifest(outputPath string, resume bool) (lettersManifest, error) {
	manifest := lettersManifest{Version: manifestVersion, Kind: kindLetters, Pages: []lettersPageResult{}}
	if !resume {
		return manifest, nil
	}
	if err := loadJSON(manifestPath(outputPath), &manifest); errors.Is(err, os.ErrNotExist) {
		return manifest, nil
	} else if err != nil {
		return manifest, fmt.Errorf("load letters-table manifest: %w", err)
	}
	if manifest.Kind != kindLetters || (manifest.Version != legacyManifestVersion && manifest.Version != twoPassManifestVersion && manifest.Version != resumableManifestVersion && manifest.Version != failureManifestVersion && manifest.Version != manifestVersion) {
		return manifest, fmt.Errorf("invalid letters-table manifest metadata in %s", manifestPath(outputPath))
	}
	migrateLettersManifest(&manifest)
	return manifest, nil
}

func migrateIndexManifest(manifest *indexManifest) {
	for i := range manifest.Pages {
		manifest.Pages[i].ImagePath = cleanPath(manifest.Pages[i].ImagePath)
	}
	oldVersion := manifest.Version
	manifest.Version = manifestVersion
	for i := range manifest.Pages {
		if manifest.Pages[i].ExtractionMode == "" {
			manifest.Pages[i].ExtractionMode = modeOnePass
		}
		if oldVersion <= twoPassManifestVersion && manifest.Pages[i].ExtractionMode == modeTwoPass {
			manifest.Pages[i].TranscriptionProvider = manifest.Pages[i].Provider
			manifest.Pages[i].TranscriptionModel = manifest.Pages[i].Model
			manifest.Pages[i].TranscribedAt = manifest.Pages[i].ExtractedAt
		}
	}
}

func migrateLettersManifest(manifest *lettersManifest) {
	for i := range manifest.Pages {
		manifest.Pages[i].ImagePath = cleanPath(manifest.Pages[i].ImagePath)
	}
	oldVersion := manifest.Version
	manifest.Version = manifestVersion
	for i := range manifest.Pages {
		if manifest.Pages[i].ExtractionMode == "" {
			manifest.Pages[i].ExtractionMode = modeOnePass
		}
		if oldVersion <= twoPassManifestVersion && manifest.Pages[i].ExtractionMode == modeTwoPass {
			manifest.Pages[i].TranscriptionProvider = manifest.Pages[i].Provider
			manifest.Pages[i].TranscriptionModel = manifest.Pages[i].Model
			manifest.Pages[i].TranscribedAt = manifest.Pages[i].ExtractedAt
		}
	}
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

func findIndexPage(pages []indexPageResult, imagePath string) *indexPageResult {
	path := cleanPath(imagePath)
	for i := range pages {
		if cleanPath(pages[i].ImagePath) == path {
			return &pages[i]
		}
	}
	return nil
}

func findLettersPage(pages []lettersPageResult, imagePath string) *lettersPageResult {
	path := cleanPath(imagePath)
	for i := range pages {
		if cleanPath(pages[i].ImagePath) == path {
			return &pages[i]
		}
	}
	return nil
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
		completed[cleanPath(page.ImagePath)] = page.Entries != nil
	}
	return completed
}

func lettersCompleted(manifest lettersManifest) map[string]bool {
	completed := make(map[string]bool, len(manifest.Pages))
	for _, page := range manifest.Pages {
		completed[cleanPath(page.ImagePath)] = page.Entries != nil
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

func cleanPath(value string) string {
	cleaned := filepath.Clean(value)
	legacyRoot := filepath.Clean("cmd/correspondence_ingest/data")
	if cleaned == legacyRoot {
		return "data"
	}
	if strings.HasPrefix(cleaned, legacyRoot+string(filepath.Separator)) {
		return filepath.Join("data", strings.TrimPrefix(cleaned, legacyRoot+string(filepath.Separator)))
	}
	return cleaned
}

func renderIndexCSV(path string, manifest indexManifest) error {
	return writeCSVAtomically(path, indexHeader, func(writer *csv.Writer) error {
		for _, page := range sortIndexPages(manifest.Pages) {
			for _, entry := range page.Entries {
				if err := writer.Write([]string{entry.Name, entry.PageNumber, entry.Reference, strconv.FormatBool(entry.IsBold), page.Volume}); err != nil {
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
	temporary, err := os.CreateTemp(filepath.Dir(path), ".correspondence_ingest-*")
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
