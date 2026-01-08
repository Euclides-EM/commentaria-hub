package filesys

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
)

func (m *Manager) RetrieveAltoPage(ann *model.Annotation, page int) (*alto.Alto, string, error) {
	pageAltoPath := filepath.Join(m.DatasetAnnotationAltoDir(ann), pagesparser.PageToXMLFilename(page))
	if _, err := os.Stat(pageAltoPath); os.IsNotExist(err) {
		return nil, pageAltoPath, fmt.Errorf("page ALTO %s does not exist for annotation %s", pageAltoPath, ann.ID)
	}

	af, err := alto.LoadFromFile(pageAltoPath)
	if err != nil {
		return nil, pageAltoPath, fmt.Errorf("load ALTO: %w", err)
	}
	return af, pageAltoPath, nil
}

func (m *Manager) ApplyToAltoPage(ann *model.Annotation, page int, applier func(*alto.Alto) error) error {
	a, filePath, err := m.RetrieveAltoPage(ann, page)
	if err != nil {
		return err
	}

	if err := applier(a); err != nil {
		return fmt.Errorf("apply to ALTO: %w", err)
	}

	if err := alto.SaveToFile(a, filePath); err != nil {
		return fmt.Errorf("save ALTO: %w", err)
	}

	return nil
}
