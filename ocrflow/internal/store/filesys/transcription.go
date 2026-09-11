package filesys

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
)

type TranscriptionFileFormat string

const (
	TranscriptionFileFormatALTO     TranscriptionFileFormat = "alto"
	TranscriptionFileFormatText     TranscriptionFileFormat = "text"
	TranscriptionFileFormatMarkdown TranscriptionFileFormat = "markdown"
)

type TranscriptionFile struct {
	Name   string
	Label  string
	Path   string
	Format TranscriptionFileFormat
}

func (m *Manager) ListEditionTranscriptionFiles(editionKey string, pageNum int) ([]TranscriptionFile, error) {
	dir := m.EditionTxtPageTranscriptionDir(editionKey, fmt.Sprintf("%d", pageNum))
	files, err := listTranscriptionFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("list edition transcription files for page %d of edition %s: %w", pageNum, editionKey, err)
	}
	return files, nil
}

func (m *Manager) ListAnnotationTranscriptionFiles(ann *annotation.Annotation, pageNumOrKey string) ([]TranscriptionFile, error) {
	dir := m.AnnotationTxtPageTranscriptionDir(ann, pageNumOrKey)
	files, err := listTranscriptionFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("list annotation transcription files for page %s of annotation %s: %w", pageNumOrKey, ann.ID, err)
	}
	return files, nil
}

func listTranscriptionFiles(dir string) ([]TranscriptionFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read transcription directory %s: %w", dir, err)
	}

	files := make([]TranscriptionFile, 0, len(entries))
	baseCounts := make(map[string]int)
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") || strings.EqualFold(entry.Name(), "METS.xml") {
			continue
		}

		format, ok := transcriptionFileFormat(entry.Name())
		if !ok {
			continue
		}
		base := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		baseCounts[strings.ToLower(base)]++
		files = append(files, TranscriptionFile{
			Name:   entry.Name(),
			Label:  base,
			Path:   filepath.Join(dir, entry.Name()),
			Format: format,
		})
	}

	for i := range files {
		base := strings.TrimSuffix(files[i].Name, filepath.Ext(files[i].Name))
		if baseCounts[strings.ToLower(base)] > 1 {
			files[i].Label = files[i].Name
		}
	}

	sort.SliceStable(files, func(i, j int) bool {
		leftOriginal := strings.EqualFold(strings.TrimSuffix(files[i].Name, filepath.Ext(files[i].Name)), "original")
		rightOriginal := strings.EqualFold(strings.TrimSuffix(files[j].Name, filepath.Ext(files[j].Name)), "original")
		if leftOriginal != rightOriginal {
			return leftOriginal
		}
		if leftOriginal {
			leftRank := transcriptionFormatRank(files[i].Format)
			rightRank := transcriptionFormatRank(files[j].Format)
			if leftRank != rightRank {
				return leftRank < rightRank
			}
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	if len(files) == 0 {
		return nil, fmt.Errorf("no ALTO, Markdown, or TXT transcription files found in %s", dir)
	}
	return files, nil
}

func transcriptionFileFormat(name string) (TranscriptionFileFormat, bool) {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".xml":
		return TranscriptionFileFormatALTO, true
	case ".md", ".markdown":
		return TranscriptionFileFormatMarkdown, true
	case ".txt":
		return TranscriptionFileFormatText, true
	default:
		return "", false
	}
}

func transcriptionFormatRank(format TranscriptionFileFormat) int {
	switch format {
	case TranscriptionFileFormatALTO:
		return 0
	case TranscriptionFileFormatMarkdown:
		return 1
	case TranscriptionFileFormatText:
		return 2
	default:
		return 3
	}
}
