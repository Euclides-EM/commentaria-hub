package transcriptioncorrector

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/formatcov"
)

func validateDirectories(cfg Config) error {
	inputDirs := append(append([]string(nil), cfg.MarkdownDirs...), cfg.ALTODirs...)
	inputDirs = append(inputDirs, cfg.TranscriptionDirs...)
	inputDirs = append(inputDirs, cfg.ImagesDir)
	for _, dir := range inputDirs {
		info, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("access input directory %s: %w", dir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("input path is not a directory: %s", dir)
		}
	}
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory %s: %w", cfg.OutputDir, err)
	}
	return nil
}

func discoverPages(imagesDir string) ([]page, error) {
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return nil, fmt.Errorf("read images directory %s: %w", imagesDir, err)
	}
	allowed := map[string]bool{".gif": true, ".jpeg": true, ".jpg": true, ".png": true, ".tif": true, ".tiff": true, ".webp": true}
	byKey := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !allowed[ext] {
			continue
		}
		key := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if !isPageKey(key) {
			continue
		}
		if old, exists := byKey[key]; exists {
			return nil, fmt.Errorf("multiple images found for %s: %s and %s", key, old, entry.Name())
		}
		byKey[key] = filepath.Join(imagesDir, entry.Name())
	}
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	if len(keys) == 0 {
		return nil, fmt.Errorf("no page-NNNN images found directly in %s", imagesDir)
	}
	pages := make([]page, 0, len(keys))
	for _, key := range keys {
		pages = append(pages, page{key: key, imagePath: byKey[key]})
	}
	return pages, nil
}

func selectPages(pages []page, pageKeys []string) ([]page, error) {
	if len(pageKeys) == 0 {
		return pages, nil
	}
	requested := make(map[string]struct{}, len(pageKeys))
	for _, key := range pageKeys {
		requested[key] = struct{}{}
	}
	selected := make([]page, 0, len(pageKeys))
	for _, p := range pages {
		if _, ok := requested[p.key]; ok {
			selected = append(selected, p)
			delete(requested, p.key)
		}
	}
	if len(requested) > 0 {
		missing := make([]string, 0, len(requested))
		for key := range requested {
			missing = append(missing, key)
		}
		slices.Sort(missing)
		return nil, fmt.Errorf("images not found for requested pages: %s", strings.Join(missing, ", "))
	}
	return selected, nil
}

func isPageKey(key string) bool {
	if !strings.HasPrefix(key, "page-") || len(key) == len("page-") {
		return false
	}
	for _, r := range key[len("page-"):] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func loadCandidates(markdownDirs, altoDirs, transcriptionDirs []string, pageKey string) ([]candidate, error) {
	result := make([]candidate, 0, len(markdownDirs)+len(altoDirs)+len(transcriptionDirs))
	for i, dir := range markdownDirs {
		path, err := markdownPath(dir, pageKey)
		if err != nil {
			return nil, err
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read transcription %s: %w", path, err)
		}
		result = append(result, candidate{
			label: fmt.Sprintf("input %d (%s)", i+1, filepath.Base(filepath.Clean(dir))),
			text:  string(contents),
		})
	}
	for i, dir := range altoDirs {
		path, err := altoPath(dir, pageKey)
		if err != nil {
			return nil, err
		}
		doc, err := alto.LoadFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("load ALTO transcription %s: %w", path, err)
		}
		result = append(result, candidate{
			label: fmt.Sprintf("ALTO input %d (%s; converted with ALTOToMarkdown)", i+1, filepath.Base(filepath.Clean(dir))),
			text:  formatcov.ALTOToMarkdown(doc),
		})
	}
	for i, dir := range transcriptionDirs {
		converted, err := loadMixedTranscriptionCandidates(dir, pageKey, i+1)
		if err != nil {
			return nil, err
		}
		result = append(result, converted...)
	}
	return result, nil
}

func loadMixedTranscriptionCandidates(dir, pageKey string, sourceIndex int) ([]candidate, error) {
	baseLabel := fmt.Sprintf("transcription input %d (%s)", sourceIndex, filepath.Base(filepath.Clean(dir)))
	var result []candidate
	for _, path := range []string{
		filepath.Join(dir, pageKey, "original.xml"),
		filepath.Join(dir, pageKey+".xml"),
	} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			doc, err := alto.LoadFromFile(path)
			if err != nil {
				return nil, fmt.Errorf("load ALTO transcription %s: %w", path, err)
			}
			result = append(result, candidate{label: baseLabel + "; ALTO converted with ALTOToMarkdown", text: formatcov.ALTOToMarkdown(doc)})
			break
		}
	}
	for _, path := range []string{
		filepath.Join(dir, pageKey, "original.md"),
		filepath.Join(dir, pageKey+".md"),
	} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			contents, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read transcription %s: %w", path, err)
			}
			result = append(result, candidate{label: baseLabel + "; Markdown", text: string(contents)})
			break
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no ALTO or Markdown transcription found for %s in %s", pageKey, dir)
	}
	return result, nil
}

func altoPath(dir, pageKey string) (string, error) {
	preferred := filepath.Join(dir, pageKey+".xml")
	if info, err := os.Stat(preferred); err == nil && !info.IsDir() {
		return preferred, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read ALTO directory %s: %w", dir, err)
	}
	var matches []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".xml") {
			continue
		}
		if strings.EqualFold(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())), pageKey) {
			matches = append(matches, filepath.Join(dir, entry.Name()))
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple ALTO XML files found for %s in %s", pageKey, dir)
	}
	return "", fmt.Errorf("no ALTO XML found for %s in %s (expected %s)", pageKey, dir, preferred)
}

func markdownPath(dir, pageKey string) (string, error) {
	preferred := filepath.Join(dir, pageKey, "original.md")
	if info, err := os.Stat(preferred); err == nil && !info.IsDir() {
		return preferred, nil
	}
	flat := filepath.Join(dir, pageKey+".md")
	if info, err := os.Stat(flat); err == nil && !info.IsDir() {
		return flat, nil
	}
	matches, err := filepath.Glob(filepath.Join(dir, pageKey, "*.md"))
	if err != nil {
		return "", fmt.Errorf("find markdown for %s in %s: %w", pageKey, dir, err)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple markdown files for %s in %s; name the intended file original.md", pageKey, dir)
	}
	return "", fmt.Errorf("no markdown found for %s in %s (expected %s or %s)", pageKey, dir, preferred, flat)
}
