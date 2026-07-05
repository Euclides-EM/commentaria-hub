package app

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
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

type letterRange struct {
	number string
	name   string
	start  int
	end    int
}

var pageRangePattern = regexp.MustCompile(`^\s*(\d+)\s*(?:-|–|—|à)\s*(\d+)`)
var firstPagePattern = regexp.MustCompile(`\d+`)

func renderIndexCSV(path string, manifest indexManifest, letters lettersManifest) error {
	ranges := buildLetterRanges(letters)
	entriesByName := indexEntriesByVolumeAndName(manifest)
	return writeCSVAtomically(path, indexHeader, func(writer *csv.Writer) error {
		for _, page := range sortIndexPages(manifest.Pages) {
			for _, entry := range page.Entries {
				pageNumbers := uniqueStrings(effectivePageNumbers(entry, page.Volume, entriesByName, map[string]bool{}))
				if len(pageNumbers) == 0 {
					pageNumbers = []string{""}
				}
				for _, effectivePage := range pageNumbers {
					matched := matchingLetterRanges([]string{effectivePage}, ranges[page.Volume])
					if len(matched) == 0 {
						matched = []*letterRange{nil}
					}
					for _, letter := range matched {
						outputPage := entry.PageNumber
						if entry.Reference != "" {
							outputPage = effectivePage
						}
						row := []string{entry.Name, outputPage, entry.Reference, strconv.FormatBool(entry.IsBold), page.Volume, "", "", "", ""}
						if letter != nil {
							row[5], row[6], row[7] = letter.number, letter.name, strconv.Itoa(letter.start)
							if letter.end > 0 {
								row[8] = strconv.Itoa(letter.end)
							}
						}
						if err := writer.Write(row); err != nil {
							return err
						}
					}
				}
			}
		}
		return nil
	})
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func buildLetterRanges(manifest lettersManifest) map[string][]letterRange {
	result := map[string][]letterRange{}
	for _, page := range sortLettersPages(manifest.Pages) {
		for _, entry := range page.Entries {
			start, err := strconv.Atoi(strings.TrimSpace(entry.PageNumber))
			if err == nil {
				result[page.Volume] = append(result[page.Volume], letterRange{number: entry.LetterNumber, name: entry.LetterName, start: start})
			}
		}
	}
	for volume := range result {
		ranges := result[volume]
		sort.SliceStable(ranges, func(i, j int) bool { return ranges[i].start < ranges[j].start })
		for i := range ranges {
			for j := i + 1; j < len(ranges); j++ {
				if ranges[j].start > ranges[i].start {
					ranges[i].end = ranges[j].start - 1
					break
				}
			}
		}
		result[volume] = ranges
	}
	return result
}

func indexEntriesByVolumeAndName(manifest indexManifest) map[string]map[string][]indexEntry {
	result := map[string]map[string][]indexEntry{}
	for _, page := range manifest.Pages {
		if result[page.Volume] == nil {
			result[page.Volume] = map[string][]indexEntry{}
		}
		for _, entry := range page.Entries {
			name := strings.TrimSpace(entry.Name)
			result[page.Volume][name] = append(result[page.Volume][name], entry)
		}
	}
	return result
}

func effectivePageNumbers(entry indexEntry, volume string, entries map[string]map[string][]indexEntry, visiting map[string]bool) []string {
	if strings.TrimSpace(entry.PageNumber) != "" {
		return []string{entry.PageNumber}
	}
	target := strings.TrimSpace(entry.Reference)
	if target == "" || visiting[target] {
		return nil
	}
	visiting[target] = true
	defer delete(visiting, target)
	var result []string
	for _, referenced := range matchingReferencedEntries(entries[volume], target) {
		result = append(result, effectivePageNumbers(referenced, volume, entries, visiting)...)
	}
	return result
}

func matchingReferencedEntries(entriesByName map[string][]indexEntry, reference string) []indexEntry {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return nil
	}
	if exact := entriesByName[reference]; len(exact) > 0 {
		return exact
	}
	var names []string
	for name := range entriesByName {
		if referenceMatchesIndexName(reference, name) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		for name := range entriesByName {
			if referenceMatchesIndexName(normalizeReferencePunctuation(reference), normalizeReferencePunctuation(name)) {
				names = append(names, name)
			}
		}
	}
	if len(names) == 0 {
		for name := range entriesByName {
			if referenceTokenMatch(reference, name) {
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)
	var result []indexEntry
	for _, name := range names {
		result = append(result, entriesByName[name]...)
	}
	return result
}

func normalizeReferencePunctuation(value string) string {
	return strings.NewReplacer("’", "'", "‘", "'", "‑", "-", "–", "-").Replace(value)
}

func referenceTokenMatch(reference, name string) bool {
	referenceTokens := normalizedNameTokens(reference)
	if len(referenceTokens) < 2 {
		return false
	}
	available := map[string]int{}
	for _, token := range normalizedNameTokens(name) {
		available[token]++
	}
	for _, token := range referenceTokens {
		if available[token] == 0 {
			return false
		}
		available[token]--
	}
	return true
}

func normalizedNameTokens(value string) []string {
	decomposed := norm.NFD.String(strings.ToLower(normalizeReferencePunctuation(value)))
	var cleaned strings.Builder
	for _, character := range decomposed {
		switch {
		case unicode.Is(unicode.Mn, character):
			continue
		case unicode.IsLetter(character) || unicode.IsDigit(character):
			cleaned.WriteRune(character)
		default:
			cleaned.WriteByte(' ')
		}
	}
	return strings.Fields(cleaned.String())
}

func referenceMatchesIndexName(reference, name string) bool {
	reference = strings.TrimSpace(reference)
	name = strings.TrimSpace(name)
	if strings.HasPrefix(name, reference) {
		return true
	}
	referenceBase, referenceQualifier, ok := splitParentheticalName(reference)
	if !ok {
		return false
	}
	nameBase, nameQualifier, ok := splitParentheticalName(name)
	return ok && nameBase == referenceBase && qualifierSuffixMatches(referenceQualifier, nameQualifier)
}

func qualifierSuffixMatches(reference, name string) bool {
	referenceParts := strings.Fields(reference)
	nameParts := strings.Fields(name)
	if len(referenceParts) > len(nameParts) {
		return false
	}
	nameParts = nameParts[len(nameParts)-len(referenceParts):]
	for i, referencePart := range referenceParts {
		if referencePart == nameParts[i] {
			continue
		}
		if !strings.HasSuffix(referencePart, ".") || !isAbbreviationOf(strings.TrimSuffix(referencePart, "."), nameParts[i]) {
			return false
		}
	}
	return true
}

func isAbbreviationOf(abbreviation, value string) bool {
	abbreviation = strings.ToLower(abbreviation)
	value = strings.ToLower(value)
	position := 0
	for _, character := range value {
		if position < len(abbreviation) && byte(character) == abbreviation[position] {
			position++
		}
	}
	return position == len(abbreviation)
}

func splitParentheticalName(value string) (string, string, bool) {
	open := strings.Index(value, "(")
	close := strings.LastIndex(value, ")")
	if open < 0 || close <= open {
		return "", "", false
	}
	base := strings.TrimSpace(value[:open])
	qualifier := strings.TrimSpace(value[open+1 : close])
	return base, qualifier, base != "" && qualifier != ""
}

func matchingLetterRanges(pageReferences []string, ranges []letterRange) []*letterRange {
	seen := map[int]bool{}
	var result []*letterRange
	for _, reference := range pageReferences {
		first, last, ok := parsePageBounds(reference)
		if !ok {
			continue
		}
		for i := range ranges {
			end := ranges[i].end
			if end == 0 {
				end = int(^uint(0) >> 1)
			}
			if ranges[i].start <= last && end >= first {
				// If several letters have the same start page, that page belongs only
				// to the last one listed in the manifest.
				if i+1 < len(ranges) && ranges[i+1].start == ranges[i].start && first <= ranges[i].start {
					continue
				}
				if !seen[i] {
					seen[i] = true
					result = append(result, &ranges[i])
				}
			}
		}
	}
	return result
}

func parsePageBounds(value string) (int, int, bool) {
	if match := pageRangePattern.FindStringSubmatch(value); match != nil {
		first, err1 := strconv.Atoi(match[1])
		last, err2 := strconv.Atoi(match[2])
		return first, last, err1 == nil && err2 == nil && first <= last
	}
	match := firstPagePattern.FindString(value)
	if match == "" {
		return 0, 0, false
	}
	page, err := strconv.Atoi(match)
	return page, page, err == nil
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
