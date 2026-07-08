package filesys

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
)

func (m *Manager) RetrieveEditionTXTPage(editionKey string, pageNum string) (lines []string, translationsByLang map[string][]string, error error) {
	dir := m.EditionTxtPageTranscriptionDir(editionKey, pageNum)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("edition TXT page transcription directory %s does not exist for edition %s", dir, editionKey)
	}

	return m.extractLinesTranscriptionAndTranslations(dir)
}

func (m *Manager) RetrieveAnnotationTXTPage(ann *annotation.Annotation, key string) (lines []string, translationsByLang map[string][]string, error error) {
	dir := m.AnnotationTxtPageTranscriptionDir(ann, key)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("edition TXT page transcription directory %s does not exist for annotation %s", dir, ann.ID)
	}

	return m.extractLinesTranscriptionAndTranslations(dir)
}

func (m *Manager) extractLinesTranscriptionAndTranslations(dir string) ([]string, map[string][]string, error) {
	linesPath := filepath.Join(dir, "original.txt")
	linesData, err := os.ReadFile(linesPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read lines TXT: %w", err)
	}
	lines := strings.Split(string(linesData), "\n")

	translationsByLang := make(map[string][]string)
	translationFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("read edition TXT page transcription directory: %w", err)
	}
	for _, file := range translationFiles {
		if file.IsDir() || !m.isTranslationFile(file.Name()) {
			continue
		}

		lang := m.extractLangFromFilename(file.Name())
		p := filepath.Join(dir, file.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, nil, fmt.Errorf("read translation TXT: %w", err)
		}
		translationsByLang[lang] = strings.Split(string(data), "\n")
	}

	return lines, translationsByLang, nil
}

func (m *Manager) isTranslationFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".txt" && filename != "original.txt"
}

func (m *Manager) extractLangFromFilename(filename string) string {
	base := filename[:len(filename)-len(filepath.Ext(filename))]
	return base
}
