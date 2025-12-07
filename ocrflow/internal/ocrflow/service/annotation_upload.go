package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/escriptorium"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/roboflow"
	"github.com/samber/lo"
	"github.com/tiendc/go-deepcopy"
	"os"
	"path"
)

type AnnotationsUploader struct {
	annotationSvc        *Annotation
	datasetSvc           *Dataset
	roboflowAPIKey       string
	pythonExecutable     string
	escriptoriumPassword string
	escriptoriumUsername string
	escriptoriumBasePath string
}

func NewAnnotationsUploader(
	annotationSvc *Annotation,
	datasetSvc *Dataset,
	roboflowAPIKey string,
	pythonExecutable string,
	escriptoriumUsername string,
	escriptoriumPassword string,
	escriptoriumBasePath string,
) *AnnotationsUploader {
	return &AnnotationsUploader{
		annotationSvc:        annotationSvc,
		datasetSvc:           datasetSvc,
		roboflowAPIKey:       roboflowAPIKey,
		pythonExecutable:     pythonExecutable,
		escriptoriumPassword: escriptoriumPassword,
		escriptoriumUsername: escriptoriumUsername,
		escriptoriumBasePath: escriptoriumBasePath,
	}
}

func (a *AnnotationsUploader) UploadToRoboflow(datasetID string, id string, rbu *model.AnnotationUploadRoboflow) (*model.Annotation, error) {
	ann, err := a.annotationSvc.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("annotation not found: %w", err)
	}
	if ann.AltoDir == "" {
		return nil, fmt.Errorf("no annotations found for roboflow upload")
	}
	if ann.YoloDir == "" {
		ann, err = a.convertAlto2Yolo(ann.DatasetID, ann.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ALTO to YOLO for roboflow upload: %w", err)
		}
	}

	params := roboflow.NewUploadDatasetParams().
		SetAPIKey(lo.Ternary(rbu.APIKey == "", a.roboflowAPIKey, rbu.APIKey)).
		SetWorkspaceID(rbu.WorkspaceID).
		SetDatasetPath(ann.YoloDir).
		SetProjectID(rbu.ProjectID).
		SetIsNotGroundTruth(rbu.IsNotGroundTruth)
	if err = roboflow.UploadDataset(a.pythonExecutable, params); err != nil {
		return nil, fmt.Errorf("failed to upload to roboflow: %w", err)
	}
	var dst *model.Annotation
	if err = deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	return dst, nil
}

func (a *AnnotationsUploader) UploadToEscriptorium(datasetID string, id string, aue *model.AnnotationUploadEscriptorium) (*model.Annotation, error) {
	ann, err := a.annotationSvc.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("annotation not found: %w", err)
	}
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset for escriptorium upload: %w", err)
	}

	client := escriptorium.NewClient(
		lo.Ternary(aue.Username == "", a.escriptoriumUsername, aue.Username),
		lo.Ternary(aue.Password == "", a.escriptoriumPassword, aue.Password),
		lo.Ternary(aue.BasePath == "", a.escriptoriumBasePath, aue.BasePath),
	)

	if err := client.Authenticate(); err != nil {
		return nil, fmt.Errorf("failed to authenticate to escriptorium: %w", err)
	}

	if ann.AltoDir == "" {
		return nil, fmt.Errorf("no ALTO annotations found for escriptorium upload")
	}

	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for escriptorium upload: %w", err)
	}
	for _, page := range pages {
		imgPath := path.Join(ds.ImagesPath, pagesparser.PageToPNGFilename(page))
		if err := client.UploadImage(aue.Document, imgPath); err != nil {
			return nil, fmt.Errorf("failed to upload image to escriptorium for page %d: %w", page, err)
		}
		altoPath := path.Join(ann.AltoDir, pagesparser.PageToXMLFilename(page))
		if err := client.UploadAnnotation(aue.Document, altoPath); err != nil {
			return nil, fmt.Errorf("failed to upload ALTO to escriptorium for page %d: %w", page, err)
		}
	}
	var dst *model.Annotation
	if err := deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	return dst, nil
}

func (a *AnnotationsUploader) convertAlto2Yolo(datasetID string, id string) (*model.Annotation, error) {
	ann, err := a.annotationSvc.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("annotation not found: %w", err)
	}
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	if ann.AltoDir == "" {
		return nil, fmt.Errorf("no ALTO annotations found for conversion")
	}
	ann.YoloDir = store.DatasetAnnotationYoloDir(ann, a.datasetSvc.dataDir)
	if err := os.RemoveAll(ann.YoloDir); err != nil {
		return nil, fmt.Errorf("failed to clear YOLO annotations dir: %w", err)
	}
	// we could use other segmonto granularities here, the options are: region, subtype, full
	// I didn't notice any difference in the output for different granularities, so I chose "full"...
	if err := formatcov.Alto2Yolo(ds.ImagesPath, ann.AltoDir, ann.YoloDir, 0, "full"); err != nil {
		return nil, fmt.Errorf("failed to convert annotations: %w", err)
	}
	var dst *model.Annotation
	if err := deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	return dst, nil
}
