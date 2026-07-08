package filesys

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/markdown"
)

func (m *Manager) RetrieveEditionMarkdownPage(editionKey string, pageNum int) (*markdown.Markdown, error) {
	dir := m.EditionTxtPageTranscriptionDir(editionKey, fmt.Sprintf("%d", pageNum))
	md, err := m.extractMarkdownPage(dir)
	if err != nil {
		return nil, fmt.Errorf("retrieve edition markdown page %d for edition %s: %w", pageNum, editionKey, err)
	}
	return md, nil
}

func (m *Manager) RetrieveAnnotationMarkdownPage(ann *annotation.Annotation, pageNum string) (*markdown.Markdown, error) {
	dir := m.AnnotationTxtPageTranscriptionDir(ann, pageNum)
	md, err := m.extractMarkdownPage(dir)
	if err != nil {
		return nil, fmt.Errorf("retrieve annotation markdown page %s for annotation %s: %w", pageNum, ann.ID, err)
	}
	return md, nil
}

func (m *Manager) extractMarkdownPage(dir string) (*markdown.Markdown, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("markdown page transcription directory %s does not exist", dir)
	}

	p := filepath.Join(dir, "original.md")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read markdown page %s: %w", p, err)
	}
	return &markdown.Markdown{Content: string(data)}, nil
}
