package filesys

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
)

func (m *Manager) RetrieveEditionAltoPage(edition *model.Edition, pageNum int) (*alto.Alto, string, error) {
	pageDir := m.EditionTxtPageTranscriptionDir(edition.Key, strconv.Itoa(pageNum))
	a, filePath, err := retrieveTranscriptionAltoPage(pageDir)
	if err != nil {
		return nil, filePath, fmt.Errorf("retrieve edition ALTO page %d for edition %s: %w", pageNum, edition.Key, err)
	}
	return a, filePath, nil
}

// RetrieveAnnotationTranscriptionAltoPage loads an ALTO transcription stored
// alongside the annotation's TXT and Markdown transcription alternatives.
func (m *Manager) RetrieveAnnotationTranscriptionAltoPage(ann *annotation.Annotation, pageNumOrKey string) (*alto.Alto, string, error) {
	pageDir := m.AnnotationTxtPageTranscriptionDir(ann, pageNumOrKey)
	a, filePath, err := retrieveTranscriptionAltoPage(pageDir)
	if err != nil {
		return nil, filePath, fmt.Errorf("retrieve annotation ALTO page %s for annotation %s: %w", pageNumOrKey, ann.ID, err)
	}
	return a, filePath, nil
}

func retrieveTranscriptionAltoPage(pageDir string) (*alto.Alto, string, error) {
	filePath := filepath.Join(pageDir, "original.xml")
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, filePath, fmt.Errorf("ALTO transcription %s does not exist", filePath)
		}
		return nil, filePath, fmt.Errorf("stat ALTO transcription %s: %w", filePath, err)
	}

	a, err := alto.LoadFromFile(filePath)
	if err != nil {
		return nil, filePath, fmt.Errorf("load ALTO transcription: %w", err)
	}
	return a, filePath, nil
}

func (m *Manager) RetrieveAnnotationAltoPage(ann *annotation.Annotation, pageNumOrKey string) (*alto.Alto, string, error) {
	var fileName string
	if page, err := strconv.Atoi(pageNumOrKey); err == nil {
		fileName = pagesparser.PageToXMLFilename(page)
	} else {
		fileName = fmt.Sprintf("%s.xml", pageNumOrKey)
	}

	pageAltoPath := filepath.Join(m.DatasetAnnotationAltoDir(ann), fileName)
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
	a, filePath, err := m.RetrieveAnnotationAltoPage(ann, fmt.Sprintf("%d", page))
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
