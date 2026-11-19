package service

import (
	"fmt"
	"log"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
)

type MetaStoreManager struct {
	datasetSvc    *Dataset
	annotationSvc *Annotation
	modelSvc      *Model
	fileSysMgt    *filesys.Manager
}

func NewMetaStoreManager(datasetSvc *Dataset, annotationSvc *Annotation, modelSvc *Model, fileSystemMgt *filesys.Manager) *MetaStoreManager {
	return &MetaStoreManager{
		datasetSvc:    datasetSvc,
		annotationSvc: annotationSvc,
		modelSvc:      modelSvc,
		fileSysMgt:    fileSystemMgt,
	}
}

func (m *MetaStoreManager) CleanupLocalStore(dryRun bool) ([]string, error) {
	dss, err := m.datasetSvc.List(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}
	log.Printf("checking local store against %d datasets...", len(dss))
	annsMap := make(map[string][]*annotation.Annotation)
	for _, ds := range dss {
		anns, err := m.annotationSvc.ListAnnotations(ds.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list annotations for dataset %s: %w", ds.ID, err)
		}
		annsMap[ds.ID] = anns
	}
	pathsToDelete, err := m.fileSysMgt.CleanupLocalStore(dryRun, annsMap, dss)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup local datasets: %w", err)
	}
	if dryRun {
		log.Printf("found %d orphaned filesystem entries during dry run", len(pathsToDelete))
	} else {
		log.Printf("cleaned up %d orphaned filesystem entries", len(pathsToDelete))
	}

	return pathsToDelete, nil
}
