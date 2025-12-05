package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/samber/lo"
	"log"
	"os"
	"path/filepath"
	"slices"
)

type MetaStoreManager struct {
	DatasetSvc     *Dataset
	AnnotationsSvc *Annotations
	ModelSvc       *Model

	ModelDir string
	DataDir  string
}

func NewMetaStoreManager(datasetSvc *Dataset, annotationsSvc *Annotations, modelSvc *Model, modelDir, dataDir string) *MetaStoreManager {
	return &MetaStoreManager{
		DatasetSvc:     datasetSvc,
		AnnotationsSvc: annotationsSvc,
		ModelSvc:       modelSvc,

		ModelDir: modelDir,
		DataDir:  dataDir,
	}
}

func (m *MetaStoreManager) CleanupLocalStore(dryRun bool) ([]string, error) {
	var toDelete []string

	dataDir, err := filepath.Abs(m.DataDir)
	if err != nil {
		return nil, fmt.Errorf("could not get abs path for data dir: %v", err)
	}

	des, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, de := range des {
		if !de.IsDir() {
			p := filepath.Join(dataDir, de.Name())
			toDelete = append(toDelete, p)
		}
		dsID := de.Name()
		if _, err := m.DatasetSvc.Get(dsID); err != nil {
			p := filepath.Join(dataDir, dsID)
			toDelete = append(toDelete, p)
			continue
		}

		ddes, err := os.ReadDir(filepath.Join(dataDir, dsID))
		if err != nil {
			return nil, fmt.Errorf("cannot read dataset dir %s: %w", dsID, err)
		}
		for _, dde := range ddes {
			if filepath.Ext(dde.Name()) == ".pdf" {
				continue
			}
			if !dde.IsDir() {
				toDelete = append(toDelete, dde.Name())
				continue
			}
			if !slices.Contains([]string{"imgs", "raw_imgs", "annotations"}, dde.Name()) {
				p := filepath.Join(dataDir, dsID, dde.Name())
				toDelete = append(toDelete, p)
			}
		}

		anns, err := m.AnnotationsSvc.ListAnnotations(dsID)
		if err != nil {
			return nil, fmt.Errorf("cannot list annotations for dataset %s: %w", dsID, err)
		}

		annsIDs := lo.Map(anns, func(ann *model.Annotation, _ int) string {
			return ann.ID
		})

		annotationsPath := filepath.Join(dataDir, dsID, "annotations")
		ddes, err = os.ReadDir(annotationsPath)
		if err != nil {
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
			log.Printf("Deleting %s", p)
			if err = os.RemoveAll(p); err != nil {
				return nil, fmt.Errorf("failed to delete %s: %w", p, err)
			}
		}
	}

	return toDelete, nil
}
