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
	"github.com/samber/lo"
	"log"
	"os"
	"path"
)

// todo: add interfaces to all services

type Annotations struct {
	pythonExecutable string
	m                map[string]*model.Annotation
	datasetSvc       *Dataset
	modelSvc         *Model
	roboflowAPIKey   string
}

func NewAnnotationsService(pythonExecutable string, datasetSvc *Dataset, modelSvc *Model, roboflowAPIKey string) *Annotations {
	return &Annotations{
		pythonExecutable: pythonExecutable,
		m:                make(map[string]*model.Annotation),
		datasetSvc:       datasetSvc,
		modelSvc:         modelSvc,
		roboflowAPIKey:   roboflowAPIKey,
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

func (a *Annotations) Create(datasetID string, ann *model.Annotation, async bool) (*model.Annotation, error) {
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

	var ocrM *model.Model
	if ann.OCRModelID() != "" {
		ocrM, err = a.modelSvc.Get(ann.OCRModelID())
		if err != nil {
			return nil, fmt.Errorf("failed to get ocr model: %w", err)
		}
	}

	ann.ID = idgen.GenerateID()
	ann.Dataset = model.Reference{ID: datasetID}
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

	errCh, err := a.execModel(ds, ann, segM, ocrM, filenames)
	if err != nil {
		return nil, err
	}

	if async {
		toUpdate := ann.DeepCopy()

		go func() {
			if recErr := <-errCh; recErr != nil {
				log.Printf("async annotation failed for %s: %v", toUpdate.ID, recErr)
				return
			}
			log.Printf("async annotation completed for %s", toUpdate.ID)

			a.m[toUpdate.ID] = toUpdate
		}()

		return nil, nil
	}

	// Synchronous path: wait for OCR to finish
	if recErr := <-errCh; recErr != nil {
		return nil, fmt.Errorf("failed to annotate facsimile: %w", recErr)
	}

	a.m[ann.ID] = ann.DeepCopy()
	return a.m[ann.ID], nil
}

func (a *Annotations) execModel(ds *model.Dataset, ann *model.Annotation, segM *model.Model, ocrM *model.Model, filenames []string) (<-chan error, error) {
	switch segM.RunWith {
	case "kraken":
		ann.AltoDir = store.DatasetAnnotationAltoDir(ann, a.datasetSvc.datasetsDir)
		return krakenwrapper.Recognize(
			ds.ImagesPath,
			ann.AltoDir,
			segM.LocalPath,
			lo.TernaryF(ocrM == nil, func() string { return "" }, func() string { return ocrM.LocalPath }),
			filenames,
		)
	case "roboflow":
		ann.RoboflowDir = store.DatasetAnnotationRoboflowDir(ann, a.datasetSvc.datasetsDir)
		return roboflow.Recognize(
			ds.ImagesPath,
			ann.RoboflowDir,
			segM.Name,
			segM.Categories,
			filenames,
			a.roboflowAPIKey,
		), nil
	}
	return nil, fmt.Errorf("unsupported model runtime: %s", segM.RunWith)
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
	if err := os.RemoveAll(ann.YoloDir); err != nil {
		return nil, fmt.Errorf("failed to clear YOLO annotations dir: %w", err)
	}
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
	params := roboflow.NewUploadDatasetParams().
		SetAPIKey(lo.Ternary(rbu.APIKey == "", a.roboflowAPIKey, rbu.APIKey)).
		SetWorkspaceID(rbu.WorkspaceID).
		SetDatasetPath(ann.YoloDir).
		SetProjectID(rbu.ProjectID).
		SetIsNotGroundTruth(rbu.IsNotGroundTruth)
	switch {
	case ann.YoloDir != "":
		params = params.SetDatasetPath(ann.YoloDir)
	case ann.RoboflowDir != "":
		params = params.SetDatasetPath(ann.RoboflowDir)
	default:
		return nil, fmt.Errorf("no annotations found for upload to roboflow")
	}
	err := roboflow.UploadDataset(a.pythonExecutable, params)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to roboflow: %w", err)
	}
	return ann.DeepCopy(), nil
}

func (a *Annotations) CreateFromZip(datasetID string, format model.AnnotationFormat, save func(dstPath string) error) (*model.Annotation, error) {
	_, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	ann := &model.Annotation{
		Meta: model.NewMeta(idgen.GenerateID()),
		//Pages             string    `json:"pages"`
		//AltoDir           string    `json:"alto_dir" readonly:"true"`
		//YoloDir           string    `json:"yolo_dir" readonly:"true"`
		Dataset: model.Reference{ID: datasetID},
		//SegmentationModel Reference `json:"segmentation_model"`
		//OCRModel          Reference `json:"ocr_model"`
	}
	var dstPath string
	switch format {
	case model.AnnotationFormatAlto:
		dstPath = store.DatasetAnnotationAltoDir(&model.Annotation{Meta: model.NewMeta(ann.ID)}, a.datasetSvc.datasetsDir)
		ann.AltoDir = dstPath
	case model.AnnotationFormatYolo:
		dstPath = store.DatasetAnnotationYoloDir(&model.Annotation{Meta: model.NewMeta(ann.ID)}, a.datasetSvc.datasetsDir)
		ann.YoloDir = dstPath
	}
	if err := save(dstPath); err != nil {
		return nil, fmt.Errorf("failed to store uploaded annotations: %w", err)
	}
	pages, err := store.InferPages(dstPath, format)
	if err != nil {
		return nil, fmt.Errorf("failed to infer pages from uploaded annotations: %w", err)
	}
	ann.Pages = pagesparser.ToString(pages)
	a.m[ann.ID] = ann.DeepCopy()
	return a.m[ann.ID], nil
}
