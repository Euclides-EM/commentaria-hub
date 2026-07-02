package filesys

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/samber/lo"
)

type Manager struct {
	baseDir          string
	modelsDir        string
	diagramsDir      string
	defaultModelsDir string
}

var protectedTopLevelStoreDirs = map[string]struct{}{
	"transcriptions": {},
}

var allowedDatasetEntries = map[string]struct{}{
	"annotations": {},
	"imgs":        {},
}

func NewFileSystemManager(baseDir, modelsDir, diagramsDir string, defaultModelsDir string) *Manager {
	return &Manager{
		baseDir:          baseDir,
		modelsDir:        modelsDir,
		diagramsDir:      diagramsDir,
		defaultModelsDir: defaultModelsDir,
	}
}

func (m *Manager) CleanupLocalStore(dryRun bool, annsMap map[string][]*annotation.Annotation, dss []*model.Dataset) ([]string, error) {
	var toDelete []string

	dataDir, err := filepath.Abs(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("could not get abs path for data dir: %v", err)
	}

	dsIDs := make([]string, 0)
	for _, ds := range dss {
		dsIDs = append(dsIDs, ds.ID)
	}

	des, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, de := range des {
		dsID := de.Name()
		if _, protected := protectedTopLevelStoreDirs[dsID]; protected {
			continue
		}
		if !de.IsDir() {
			p := filepath.Join(dataDir, de.Name())
			toDelete = append(toDelete, p)
			continue
		}
		if !slices.Contains(dsIDs, dsID) {
			p := filepath.Join(dataDir, dsID)
			toDelete = append(toDelete, p)
			continue
		}

		ddes, err := os.ReadDir(filepath.Join(dataDir, dsID))
		if err != nil {
			return nil, fmt.Errorf("cannot read dataset dir %s: %w", dsID, err)
		}
		for _, dde := range ddes {
			if _, allowed := allowedDatasetEntries[dde.Name()]; allowed {
				continue
			}
			if filepath.Ext(dde.Name()) == ".pdf" {
				p := filepath.Join(dataDir, dsID, dde.Name())
				toDelete = append(toDelete, p)
				continue
			}
			if !dde.IsDir() {
				p := filepath.Join(dataDir, dsID, dde.Name())
				toDelete = append(toDelete, p)
				continue
			}
			if !slices.Contains([]string{"imgs", "annotations"}, dde.Name()) {
				p := filepath.Join(dataDir, dsID, dde.Name())
				toDelete = append(toDelete, p)
			}
		}

		annsIDs := lo.Map(annsMap[dsID], func(ann *annotation.Annotation, _ int) string {
			return ann.ID
		})

		annotationsPath := filepath.Join(dataDir, dsID, "annotations")
		ddes, err = os.ReadDir(annotationsPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot read dataset dir %s: %w", dsID, err)
		}

		for _, dde := range ddes {
			if !dde.IsDir() {
				p := filepath.Join(annotationsPath, dde.Name())
				toDelete = append(toDelete, p)
				continue
			}
			annID := dde.Name()
			found := false
			for _, existingAnnID := range annsIDs {
				if annID == existingAnnID {
					found = true
					break
				}
			}
			if !found {
				p := filepath.Join(annotationsPath, annID)
				toDelete = append(toDelete, p)
			}
		}
	}

	if !dryRun {
		for _, p := range toDelete {
			if err = ensureStorePath(dataDir, p); err != nil {
				return nil, err
			}
			log.Printf("Deleting %s", p)
			if err = os.RemoveAll(p); err != nil {
				return nil, fmt.Errorf("failed to delete %s: %w", p, err)
			}
		}
	}

	return toDelete, nil
}

func ensureStorePath(dataDir, p string) error {
	if !filepath.IsAbs(p) {
		return fmt.Errorf("refusing to delete non-absolute path %s", p)
	}

	absPath, err := filepath.Abs(p)
	if err != nil {
		return fmt.Errorf("could not get abs path for %s: %w", p, err)
	}

	rel, err := filepath.Rel(dataDir, absPath)
	if err != nil {
		return fmt.Errorf("could not validate path %s against data dir %s: %w", p, dataDir, err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("refusing to delete path outside data dir: %s", p)
	}
	return nil
}

func (m *Manager) DeleteDatasetFiles(ds *model.Dataset) error {
	dsPath := m.DatasetDir(ds.ID)
	if _, err := os.Stat(dsPath); os.IsNotExist(err) {
		return nil
	}
	if err := os.RemoveAll(dsPath); err != nil {
		return fmt.Errorf("failed to delete dataset files at %s: %w", dsPath, err)
	}
	return nil
}
