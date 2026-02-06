package ocrflow

import (
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
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
	annsMap := make(map[string][]*ocrflow.Annotation)
	for _, ds := range dss {
		anns, err := m.annotationSvc.ListAnnotations(ds.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list annotations for dataset %s: %w", ds.ID, err)
		}
		annsMap[ds.ID] = anns
	}
	return m.fileSysMgt.CleanupLocalStore(dryRun, annsMap, dss)
}
