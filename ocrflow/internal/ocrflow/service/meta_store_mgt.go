package service

import (
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
)

type MetaStoreManager struct {
	datasetSvc    *Dataset
	annotationSvc *Annotation
	modelSvc      *Model
	fileSysMgt    *store.FileSystemManager
}

func NewMetaStoreManager(datasetSvc *Dataset, annotationSvc *Annotation, modelSvc *Model, fileSystemMgt *store.FileSystemManager) *MetaStoreManager {
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
	annsMap := make(map[string][]*model.Annotation)
	for _, ds := range dss {
		anns, err := m.annotationSvc.ListAnnotations(ds.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list annotations for dataset %s: %w", ds.ID, err)
		}
		annsMap[ds.ID] = anns
	}
	return m.fileSysMgt.CleanupLocalStore(dryRun, annsMap, dss)
}
