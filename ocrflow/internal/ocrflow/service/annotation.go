package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/krakenwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"os"
	"path"
)

// todo: add interfaces to all services

type Annotations struct {
	m          map[string]*model.Annotation
	datasetSvc *Dataset
	modelSvc   *Model
}

func NewAnnotationsService(datasetSvc *Dataset, modelSvc *Model) *Annotations {
	return &Annotations{
		m:          make(map[string]*model.Annotation),
		datasetSvc: datasetSvc,
		modelSvc:   modelSvc,
	}
}

func (a *Annotations) ListAnnotations(id string) ([]*model.Annotation, error) {
	annotations := make([]*model.Annotation, 0)
	for _, annotation := range a.m {
		if annotation.DatasetID() == id {
			annotations = append(annotations, annotation.DeepCopy())
		}
	}
	return annotations, nil
}

func (a *Annotations) Create(datasetID string, ann *model.Annotation) (*model.Annotation, error) {
	// todo: this must be async with job status tracking
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	if ds.ImagesPath == "" {
		return nil, fmt.Errorf("no JPGs path found for dataset")
	}

	m, err := a.modelSvc.Get(ann.ModelID())
	if err != nil {
		return nil, fmt.Errorf("failed to get kraken model: %w", err)
	}
	if m.KrakenRef == "" {
		return nil, fmt.Errorf("no kraken model name found for model")
	}

	ann.ID = idgen.GenerateID()
	ann.Dataset = model.Reference{ID: datasetID}
	ann.LocalDir = store.DatasetAnnotationsPath(ann, a.datasetSvc.datasetsDir)
	var filenames []string
	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages: %w", err)
	}
	for _, page := range pages {
		filename := fmt.Sprintf("%04d.jpg", page)
		if _, err := os.Stat(path.Join(ds.ImagesPath, filename)); err != nil {
			return nil, fmt.Errorf("no such file %s in existing dataset", filename)
		}
		filenames = append(filenames, filename)
	}

	if err := krakenwrapper.Recognize(ds.ImagesPath, ann.LocalDir, m.KrakenRef, filenames); err != nil {
		return nil, fmt.Errorf("failed to annotate facsimile: %w", err)
	}
	a.m[ann.ID] = ann.DeepCopy()
	return a.m[ann.ID], nil
}
