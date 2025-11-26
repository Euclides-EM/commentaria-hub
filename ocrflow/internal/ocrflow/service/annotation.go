package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/krakenwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/roboflow"
	"os"
	"path"
)

// todo: add interfaces to all services

type Annotations struct {
	pythonExecutable string
	m                map[string]*model.Annotation
	datasetSvc       *Dataset
	modelSvc         *Model
}

func NewAnnotationsService(pythonExecutable string, datasetSvc *Dataset, modelSvc *Model) *Annotations {
	return &Annotations{
		pythonExecutable: pythonExecutable,
		m:                make(map[string]*model.Annotation),
		datasetSvc:       datasetSvc,
		modelSvc:         modelSvc,
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

	segM, err := a.modelSvc.Get(ann.SegmentationModelID())
	if err != nil {
		return nil, fmt.Errorf("failed to get kraken model: %w", err)
	}

	ocrM, err := a.modelSvc.Get(ann.OCRModelID())
	if err != nil {
		return nil, fmt.Errorf("failed to get ocr model: %w", err)
	}

	ann.ID = idgen.GenerateID()
	ann.Dataset = model.Reference{ID: datasetID}
	ann.AltoDir = store.DatasetAnnotationAltoDir(ann, a.datasetSvc.datasetsDir)
	var filenames []string
	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages: %w", err)
	}
	for _, page := range pages {
		filename := fmt.Sprintf("page-%04d.png", page)
		if _, err := os.Stat(path.Join(ds.ImagesPath, filename)); err != nil {
			return nil, fmt.Errorf("no such file %s in existing dataset", filename)
		}
		filenames = append(filenames, filename)
	}

	if err := krakenwrapper.Recognize(ds.ImagesPath, ann.AltoDir, segM.LocalPath, ocrM.LocalPath, filenames); err != nil {
		return nil, fmt.Errorf("failed to annotate facsimile: %w", err)
	}
	a.m[ann.ID] = ann.DeepCopy()
	return a.m[ann.ID], nil
}

func (a *Annotations) Convert(datasetID string, id string, annc *model.AnnotationConvert) (*model.Annotation, error) {
	if annc.From != model.AnnotationFormatAlto || annc.To != model.AnnotationFormatYolo {
		return nil, fmt.Errorf("unsupported annotation format for conversion %s -> %s", annc.From, annc.To)
	}
	ann, ok := a.m[id]
	if !ok || ann.DatasetID() != datasetID {
		return nil, fmt.Errorf("annotation not found")
	}
	if ann.AltoDir == "" {
		return nil, fmt.Errorf("no ALTO annotations found for conversion")
	}
	ann.YoloDir = store.DatasetAnnotationYoloDir(ann, a.datasetSvc.datasetsDir)
	if err := formatcov.Alto2Yolo(ann.AltoDir, ann.YoloDir, annc.Shuffle, string(annc.SegmontoGranularity)); err != nil {
		return nil, fmt.Errorf("failed to convert annotations: %w", err)
	}
	return ann.DeepCopy(), nil
}

func (a *Annotations) UploadToRoboflow(datasetID string, id string, rbu *model.AnnotationRoboflowUpload) (*model.Annotation, error) {
	ann, ok := a.m[id]
	if !ok || ann.DatasetID() != datasetID {
		return nil, fmt.Errorf("annotation not found")
	}
	if ann.YoloDir == "" {
		return nil, fmt.Errorf("no YOLO annotations found for upload")
	}
	err := roboflow.UploadDataset(a.pythonExecutable, roboflow.NewUploadDatasetParams().
		SetAPIKey(rbu.APIKey).
		SetWorkspaceID(rbu.WorkspaceID).
		SetDatasetPath(ann.YoloDir).
		SetProjectID(rbu.ProjectID).
		SetIsNotGroundTruth(rbu.IsNotGroundTruth),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to roboflow: %w", err)
	}
	return ann.DeepCopy(), nil
}
