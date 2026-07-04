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

const manifestVersion = 4

const legacyManifestVersion = 1
const twoPassManifestVersion = 2
const resumableManifestVersion = 3

const responseValidationAttempts = 2

var (
	indexHeader   = []string{"name", "page_number", "reference", "is_bold", "volume"}
	lettersHeader = []string{"letter_number", "letter_name", "page_number", "volume"}
)

type indexManifest struct {
	Version int               `json:"version"`
	Kind    string            `json:"kind"`
	Pages   []indexPageResult `json:"pages"`
}

type indexPageResult struct {
	ImagePath             string       `json:"image_path"`
	Volume                string       `json:"volume"`
	Provider              string       `json:"provider"`
	Model                 string       `json:"model"`
	ExtractedAt           time.Time    `json:"extracted_at"`
	ExtractionMode        string       `json:"extraction_mode"`
	Transcription         string       `json:"transcription,omitempty"`
	TranscriptionProvider string       `json:"transcription_provider,omitempty"`
	TranscriptionModel    string       `json:"transcription_model,omitempty"`
	TranscribedAt         time.Time    `json:"transcribed_at,omitempty"`
	Entries               []indexEntry `json:"entries"`
	Issues                []string     `json:"issues,omitempty"`
	Failure               *pageFailure `json:"failure,omitempty"`
}

type lettersManifest struct {
	Version int                 `json:"version"`
	Kind    string              `json:"kind"`
	Pages   []lettersPageResult `json:"pages"`
}

type lettersPageResult struct {
	ImagePath             string        `json:"image_path"`
	Volume                string        `json:"volume"`
	Provider              string        `json:"provider"`
	Model                 string        `json:"model"`
	ExtractedAt           time.Time     `json:"extracted_at"`
	ExtractionMode        string        `json:"extraction_mode"`
	Transcription         string        `json:"transcription,omitempty"`
	TranscriptionProvider string        `json:"transcription_provider,omitempty"`
	TranscriptionModel    string        `json:"transcription_model,omitempty"`
	TranscribedAt         time.Time     `json:"transcribed_at,omitempty"`
	Entries               []letterEntry `json:"entries"`
	Issues                []string      `json:"issues,omitempty"`
	Failure               *pageFailure  `json:"failure,omitempty"`
}

type pageFailure struct {
	Phase    string    `json:"phase"`
	Provider string    `json:"provider"`
	Model    string    `json:"model"`
	Error    string    `json:"error"`
	FailedAt time.Time `json:"failed_at"`
}

const (
	failurePhaseSingle = "single-pass"
	failurePhaseFirst  = "first-pass"
	failurePhaseSecond = "second-pass"
)

type checkpointError struct{ err error }

func (e checkpointError) Error() string { return e.err.Error() }
func (e checkpointError) Unwrap() error { return e.err }

type pageIssues struct {
	ImagePath string
	Issues    []string
}

type failedPage struct {
	ImagePath string
	Failure   *pageFailure
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
		existing := findIndexPage(manifest.Pages, image.Path)
		if cfg.resume && !targeted && (completed[cleanPath(image.Path)] || (cfg.skipFailures && existing != nil && existing.Failure != nil)) {
			skipped++
			reason := "completed"
			if !completed[cleanPath(image.Path)] {
				reason = "failed"
			}
			fmt.Fprintf(out, "[%d/%d] Skipping %s index page %s\n", i+1, len(selected), reason, image.Path)
			continue
		}
		fmt.Fprintf(out, "[%d/%d] Extracting index page %s\n", i+1, len(selected), image.Path)
		cached := ""
		if existing != nil && effectiveExtractionMode(cfg) == modeTwoPass {
			cached = existing.Transcription
		}
		entries, issues, transcription, err := extractIndexEntriesWithCheckpoint(cfg, client, image, cached, out, func(value string) error {
			manifest.Pages = upsertIndexPage(manifest.Pages, indexPageResult{
				ImagePath: cleanPath(image.Path), Volume: image.Volume, ExtractionMode: modeTwoPass,
				Transcription: value, TranscriptionProvider: firstPassProvider(cfg), TranscriptionModel: firstPassModel(cfg), TranscribedAt: time.Now().UTC(),
			})
			if err := saveJSONAtomically(manifestPath(cfg.indexCSV), manifest); err != nil {
				return err
			}
			return renderIndexCSV(cfg.indexCSV, manifest)
		})
		if err != nil {
			var checkpointErr checkpointError
			if errors.As(err, &checkpointErr) {
				return err
			}
			failure := newPageFailure(cfg, transcription, err)
			page := findIndexPage(manifest.Pages, image.Path)
			if page == nil {
				manifest.Pages = upsertIndexPage(manifest.Pages, indexPageResult{ImagePath: cleanPath(image.Path), Volume: image.Volume, ExtractionMode: effectiveExtractionMode(cfg), Failure: failure})
			} else {
				page.Failure = failure
			}
			if err := saveJSONAtomically(manifestPath(cfg.indexCSV), manifest); err != nil {
				return fmt.Errorf("save index failure after %s: %w", image.Path, err)
			}
			if err := renderIndexCSV(cfg.indexCSV, manifest); err != nil {
				return fmt.Errorf("render index CSV after failure %s: %w", image.Path, err)
			}
			fmt.Fprintf(out, "Warning: index extraction failed for %s during %s: %v; continuing\n", image.Path, failure.Phase, err)
			continue
		}
		for _, issue := range issues {
			fmt.Fprintf(out, "Warning: index response for %s: %s; skipping entry\n", image.Path, issue)
		}
		manifest.Pages = upsertIndexPage(manifest.Pages, indexPageResult{
			ImagePath: cleanPath(image.Path), Volume: image.Volume, Provider: extractionProvider(cfg),
			Model: extractionModel(cfg), ExtractedAt: time.Now().UTC(), ExtractionMode: effectiveExtractionMode(cfg),
			Transcription: transcription, Entries: entries, Issues: issues,
		})
		if effectiveExtractionMode(cfg) == modeTwoPass {
			page := findIndexPage(manifest.Pages, image.Path)
			page.TranscriptionProvider, page.TranscriptionModel = firstPassProvider(cfg), firstPassModel(cfg)
			if page.TranscribedAt.IsZero() {
				page.TranscribedAt = time.Now().UTC()
			}
		}
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

func extractIndexEntries(cfg config, client llmExecutor, image imageInput, out io.Writer) ([]indexEntry, error) {
	entries, _, err := extractIndexEntriesWithIssues(cfg, client, image, out)
	return entries, err
}

func extractIndexEntriesWithIssues(cfg config, client llmExecutor, image imageInput, out io.Writer) ([]indexEntry, []string, error) {
	entries, issues, _, err := extractIndexEntriesWithAudit(cfg, client, image, out)
	return entries, issues, err
}

func extractIndexEntriesWithAudit(cfg config, client llmExecutor, image imageInput, out io.Writer) ([]indexEntry, []string, string, error) {
	return extractIndexEntriesWithCheckpoint(cfg, client, image, "", out, nil)
}

func extractIndexEntriesWithCheckpoint(cfg config, client llmExecutor, image imageInput, cached string, out io.Writer, checkpoint func(string) error) ([]indexEntry, []string, string, error) {
	transcription, prompt, attachment, err := extractionInput(cfg, client, image, indexPrompt, cached, checkpoint)
	if err != nil {
		return nil, nil, transcription, fmt.Errorf("transcribe index image %s: %w", image.Path, err)
	}
	var parseErr error
	for attempt := 1; attempt <= responseValidationAttempts; attempt++ {
		raw, err := client.Exec(extractionProvider(cfg), extractionModel(cfg), prompt, attachment)
		if err != nil {
			return nil, nil, transcription, fmt.Errorf("extract index image %s: %w", image.Path, err)
		}
		entries, issues, err := parseIndexResponseWithIssues(raw, image.Volume)
		if err == nil {
			return entries, issues, transcription, nil
		}
		parseErr = err
		if attempt < responseValidationAttempts {
			fmt.Fprintf(out, "Invalid index response for %s (%v); retrying extraction (%d/%d)\n", image.Path, err, attempt+1, responseValidationAttempts)
		}
	}
	return nil, nil, transcription, fmt.Errorf("parse index response for %s after %d attempts: %w", image.Path, responseValidationAttempts, parseErr)
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
		existing := findLettersPage(manifest.Pages, image.Path)
		if cfg.resume && !targeted && (completed[cleanPath(image.Path)] || (cfg.skipFailures && existing != nil && existing.Failure != nil)) {
			skipped++
			reason := "completed"
			if !completed[cleanPath(image.Path)] {
				reason = "failed"
			}
			fmt.Fprintf(out, "[%d/%d] Skipping %s letters-table page %s\n", i+1, len(selected), reason, image.Path)
			continue
		}
		fmt.Fprintf(out, "[%d/%d] Extracting letters-table page %s\n", i+1, len(selected), image.Path)
		cached := ""
		if existing != nil && effectiveExtractionMode(cfg) == modeTwoPass {
			cached = existing.Transcription
		}
		transcription, prompt, attachment, err := extractionInput(cfg, client, image, lettersTablePrompt, cached, func(value string) error {
			manifest.Pages = upsertLettersPage(manifest.Pages, lettersPageResult{
				ImagePath: cleanPath(image.Path), Volume: image.Volume, ExtractionMode: modeTwoPass,
				Transcription: value, TranscriptionProvider: firstPassProvider(cfg), TranscriptionModel: firstPassModel(cfg), TranscribedAt: time.Now().UTC(),
			})
			if err := saveJSONAtomically(manifestPath(cfg.lettersCSV), manifest); err != nil {
				return err
			}
			return renderLettersCSV(cfg.lettersCSV, manifest)
		})
		if err != nil {
			var checkpointErr checkpointError
			if errors.As(err, &checkpointErr) {
				return err
			}
			err = fmt.Errorf("transcribe letters-table image %s: %w", image.Path, err)
			if persistErr := recordLettersFailure(cfg, &manifest, image, transcription, err); persistErr != nil {
				return persistErr
			}
			fmt.Fprintf(out, "Warning: letters-table extraction failed for %s during %s: %v; continuing\n", image.Path, newPageFailure(cfg, transcription, err).Phase, err)
			continue
		}
		raw, err := client.Exec(extractionProvider(cfg), extractionModel(cfg), prompt, attachment)
		if err != nil {
			err = fmt.Errorf("extract letters-table image %s: %w", image.Path, err)
			if persistErr := recordLettersFailure(cfg, &manifest, image, transcription, err); persistErr != nil {
				return persistErr
			}
			fmt.Fprintf(out, "Warning: letters-table extraction failed for %s during %s: %v; continuing\n", image.Path, newPageFailure(cfg, transcription, err).Phase, err)
			continue
		}
		entries, issues, err := parseLettersResponseWithIssues(raw, image.Volume)
		if err != nil {
			err = fmt.Errorf("parse letters-table response for %s: %w", image.Path, err)
			if persistErr := recordLettersFailure(cfg, &manifest, image, transcription, err); persistErr != nil {
				return persistErr
			}
			fmt.Fprintf(out, "Warning: letters-table extraction failed for %s during %s: %v; continuing\n", image.Path, newPageFailure(cfg, transcription, err).Phase, err)
			continue
		}
		for _, issue := range issues {
			fmt.Fprintf(out, "Warning: letters-table response for %s: %s; skipping entry\n", image.Path, issue)
		}
		manifest.Pages = upsertLettersPage(manifest.Pages, lettersPageResult{
			ImagePath: cleanPath(image.Path), Volume: image.Volume, Provider: extractionProvider(cfg),
			Model: extractionModel(cfg), ExtractedAt: time.Now().UTC(), ExtractionMode: effectiveExtractionMode(cfg),
			Transcription: transcription, Entries: entries, Issues: issues,
		})
		if effectiveExtractionMode(cfg) == modeTwoPass {
			page := findLettersPage(manifest.Pages, image.Path)
			page.TranscriptionProvider, page.TranscriptionModel = firstPassProvider(cfg), firstPassModel(cfg)
			if page.TranscribedAt.IsZero() {
				page.TranscribedAt = time.Now().UTC()
			}
		}
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

func effectiveExtractionMode(cfg config) string {
	if cfg.extractionMode == modeTwoPass {
		return modeTwoPass
	}
	return modeOnePass
}

func extractionInput(cfg config, client llmExecutor, image imageInput, structuredPrompt, cached string, checkpoint func(string) error) (transcription, prompt, attachment string, err error) {
	if effectiveExtractionMode(cfg) == modeOnePass {
		return "", structuredPrompt, image.Path, nil
	}
	transcription = cached
	if strings.TrimSpace(transcription) == "" {
		transcription, err = client.Exec(firstPassProvider(cfg), firstPassModel(cfg), transcriptionPrompt, image.Path)
		if err != nil {
			return "", "", "", err
		}
		if strings.TrimSpace(transcription) == "" {
			return "", "", "", errors.New("LLM returned an empty transcription")
		}
		if checkpoint != nil {
			if err := checkpoint(transcription); err != nil {
				return transcription, "", "", checkpointError{fmt.Errorf("checkpoint transcription: %w", err)}
			}
		}
	}
	prompt = structuredPrompt + "\n\nParse the following transcription. It is the only source; do not infer text that is absent. Markdown **bold** markers indicate bold print.\n\n<transcription>\n" + transcription + "\n</transcription>"
	return transcription, prompt, "", nil
}

func newPageFailure(cfg config, transcription string, err error) *pageFailure {
	phase, provider, model := failurePhaseSingle, extractionProvider(cfg), extractionModel(cfg)
	if effectiveExtractionMode(cfg) == modeTwoPass {
		phase, provider, model = failurePhaseFirst, firstPassProvider(cfg), firstPassModel(cfg)
		if strings.TrimSpace(transcription) != "" {
			phase, provider, model = failurePhaseSecond, secondPassProvider(cfg), secondPassModel(cfg)
		}
	}
	return &pageFailure{Phase: phase, Provider: provider, Model: model, Error: err.Error(), FailedAt: time.Now().UTC()}
}

func recordLettersFailure(cfg config, manifest *lettersManifest, image imageInput, transcription string, failureErr error) error {
	failure := newPageFailure(cfg, transcription, failureErr)
	page := findLettersPage(manifest.Pages, image.Path)
	if page == nil {
		manifest.Pages = upsertLettersPage(manifest.Pages, lettersPageResult{ImagePath: cleanPath(image.Path), Volume: image.Volume, ExtractionMode: effectiveExtractionMode(cfg), Failure: failure})
	} else {
		page.Failure = failure
	}
	if err := saveJSONAtomically(manifestPath(cfg.lettersCSV), *manifest); err != nil {
		return fmt.Errorf("save letters-table failure after %s: %w", image.Path, err)
	}
	if err := renderLettersCSV(cfg.lettersCSV, *manifest); err != nil {
		return fmt.Errorf("render letters-table CSV after failure %s: %w", image.Path, err)
	}
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
	if manifest.Kind != kindIndex || (manifest.Version != legacyManifestVersion && manifest.Version != twoPassManifestVersion && manifest.Version != resumableManifestVersion && manifest.Version != manifestVersion) {
		return manifest, fmt.Errorf("invalid index manifest metadata in %s", manifestPath(outputPath))
	}
	if manifest.Version != manifestVersion {
		migrateIndexManifest(&manifest)
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
	if manifest.Kind != kindLetters || (manifest.Version != legacyManifestVersion && manifest.Version != twoPassManifestVersion && manifest.Version != resumableManifestVersion && manifest.Version != manifestVersion) {
		return manifest, fmt.Errorf("invalid letters-table manifest metadata in %s", manifestPath(outputPath))
	}
	if manifest.Version != manifestVersion {
		migrateLettersManifest(&manifest)
	}
	return manifest, nil
}

func migrateIndexManifest(manifest *indexManifest) {
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

func cleanPath(path string) string { return filepath.Clean(path) }

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
			reportIssues("index", indexManifestIssues(manifest), out)
			reportFailures("index", indexFailures(manifest), out)
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
			images, err := discoverImages(cfg.lettersDir)
			if err != nil {
				return fmt.Errorf("validate completed letters-table volumes: %w", err)
			}
			sequenceErr := validateCompletedLetterVolumes(manifest, images)
			if err := validateLettersCSVMatchesManifest(cfg.lettersCSV, manifest); err != nil {
				return err
			}
			if sequenceErr == nil {
				fmt.Fprintf(out, "Valid letters-table dataset: %d rows from %d images in %s\n", rows, len(manifest.Pages), cfg.lettersCSV)
			} else {
				fmt.Fprintf(out, "Letters-table dataset structure valid: %d rows from %d images in %s\n", rows, len(manifest.Pages), cfg.lettersCSV)
			}
			reportIssues("letters", lettersManifestIssues(manifest), out)
			reportFailures("letters", lettersFailures(manifest), out)
			if sequenceErr != nil {
				return sequenceErr
			}
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
		}
		rows += len(page.Entries)
	}
	if rows != csvRows {
		return fmt.Errorf("validate letters-table dataset: manifest has %d entries but CSV has %d rows", rows, csvRows)
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
		reportIssues("index", indexManifestIssues(manifest), out)
		reportFailures("index", indexFailures(manifest), out)
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
		reportIssues("letters", lettersManifestIssues(manifest), out)
		reportFailures("letters", lettersFailures(manifest), out)
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

func indexManifestIssues(manifest indexManifest) []pageIssues {
	pages := make([]pageIssues, 0)
	for _, page := range manifest.Pages {
		if len(page.Issues) > 0 {
			pages = append(pages, pageIssues{ImagePath: page.ImagePath, Issues: page.Issues})
		}
	}
	return pages
}

func lettersManifestIssues(manifest lettersManifest) []pageIssues {
	pages := make([]pageIssues, 0)
	for _, page := range manifest.Pages {
		if len(page.Issues) > 0 {
			pages = append(pages, pageIssues{ImagePath: page.ImagePath, Issues: page.Issues})
		}
	}
	return pages
}

func reportIssues(label string, pages []pageIssues, out io.Writer) {
	total := 0
	for _, page := range pages {
		total += len(page.Issues)
	}
	fmt.Fprintf(out, "%s: %d tolerated parsing issues across %d affected images\n", label, total, len(pages))
	for _, page := range pages {
		for _, issue := range page.Issues {
			fmt.Fprintf(out, "  %s: %s\n", page.ImagePath, issue)
		}
	}
}

func indexFailures(manifest indexManifest) []failedPage {
	pages := make([]failedPage, 0)
	for _, page := range manifest.Pages {
		if page.Failure != nil {
			pages = append(pages, failedPage{ImagePath: page.ImagePath, Failure: page.Failure})
		}
	}
	return pages
}

func lettersFailures(manifest lettersManifest) []failedPage {
	pages := make([]failedPage, 0)
	for _, page := range manifest.Pages {
		if page.Failure != nil {
			pages = append(pages, failedPage{ImagePath: page.ImagePath, Failure: page.Failure})
		}
	}
	return pages
}

func reportFailures(label string, pages []failedPage, out io.Writer) {
	fmt.Fprintf(out, "%s: %d failed images\n", label, len(pages))
	for _, page := range pages {
		fmt.Fprintf(out, "  %s: %s via %s/%s: %s\n", page.ImagePath, page.Failure.Phase, page.Failure.Provider, page.Failure.Model, page.Failure.Error)
	}
}
