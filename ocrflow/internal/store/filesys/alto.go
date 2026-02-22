package filesys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
)

func (m *Manager) RetrieveEditionAltoPage(edition *model.Edition, pageNum int) (*alto.Alto, string, error) {
	return nil, "", errors.New("edition ALTO retrieval not implemented yet")
}

func (m *Manager) RetrieveAnnotationAltoPage(ann *annotation.Annotation, page int) (*alto.Alto, string, error) {
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

func (m *Manager) ApplyToAltoPage(ann *annotation.Annotation, page int, applier func(*alto.Alto) error) error {
	a, filePath, err := m.RetrieveAnnotationAltoPage(ann, page)
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
