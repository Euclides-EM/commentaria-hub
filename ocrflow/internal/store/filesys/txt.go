package filesys

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
)

func (m *Manager) RetrieveEditionTXTPage(edition *model.Edition, pageNum int) (lines []string, translationsByLang map[string][]string, error error) {
	dir := m.EditionTxtPageTranscriptionDir(edition, pageNum)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("edition TXT page transcription directory %s does not exist for edition %s", dir, edition.Key)
	}

	linesPath := filepath.Join(dir, "original.txt")
	linesData, err := os.ReadFile(linesPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read lines TXT: %w", err)
	}
	lines = formatcov.SplitLines(string(linesData))

	translationsByLang = make(map[string][]string)
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
		translationsByLang[lang] = formatcov.SplitLines(string(data))
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
