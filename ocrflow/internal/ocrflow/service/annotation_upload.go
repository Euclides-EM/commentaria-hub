package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/escriptorium"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/roboflow"
	"github.com/samber/lo"
	"path"
)

func (a *Annotations) UploadToRoboflow(datasetID string, id string, rbu *model.AnnotationUploadRoboflow) (*model.Annotation, error) {
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
	annotationDirToUpload, err := a.annotationDirForRoboflowUpload(ann)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation dir for roboflow upload: %w", err)
	}
	params = params.SetDatasetPath(annotationDirToUpload)
	if err = roboflow.UploadDataset(a.pythonExecutable, params); err != nil {
		return nil, fmt.Errorf("failed to upload to roboflow: %w", err)
	}
	return ann.DeepCopy(), nil
}

func (a *Annotations) annotationDirForRoboflowUpload(ann *model.Annotation) (string, error) {
	if ann.YoloDir != "" {
		return ann.YoloDir, nil
	}
	if ann.RoboflowDir != "" {
		return ann.RoboflowDir, nil
	}
	if ann.AltoDir == "" {
		return "", fmt.Errorf("no annotations found for roboflow upload")
	}
	_, err := a.Convert(ann.DatasetID(), ann.ID, &model.AnnotationConvert{
		From: model.AnnotationFormatAlto,
		To:   model.AnnotationFormatYolo,
	})
	if err != nil {
		return "", fmt.Errorf("failed to convert ALTO to YOLO for roboflow upload: %w", err)
	}
	return ann.YoloDir, nil
}

func (a *Annotations) UploadToEscriptorium(datasetID string, id string, aue *model.AnnotationUploadEscriptorium) (*model.Annotation, error) {
	ann, ok := a.m[id]
	if !ok || ann.DatasetID() != datasetID {
		return nil, fmt.Errorf("annotation not found")
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
		if ann.YoloDir != "" {
			var err error
			ann, err = a.Convert(ann.DatasetID(), ann.ID, &model.AnnotationConvert{
				From: model.AnnotationFormatYolo,
				To:   model.AnnotationFormatAlto,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to convert YOLO to ALTO for escriptorium upload: %w", err)
			}
		}
	}

	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for escriptorium upload: %w", err)
	}
	for _, page := range pages {
		imgPath := path.Join(ann.AltoDir, fmt.Sprintf("page-%04d.png", page))
		if err := client.UploadImage(aue.Document, imgPath); err != nil {
			return nil, fmt.Errorf("failed to upload image to escriptorium for page %d: %w", page, err)
		}
		altoPath := path.Join(ann.AltoDir, fmt.Sprintf("page-%04d.xml", page))
		if err := client.UploadAnnotation(aue.Document, altoPath); err != nil {
			return nil, fmt.Errorf("failed to upload ALTO to escriptorium for page %d: %w", page, err)
		}
	}
	return ann.DeepCopy(), nil
}

func (a *Annotations) annotationDirForEscriptoriumUpload(ann *model.Annotation) (string, error) {
	if ann.AltoDir != "" {
		return ann.AltoDir, nil
	}
	if ann.YoloDir == "" {
		return "", fmt.Errorf("no annotations found for escriptorium upload")
	}
	_, err := a.Convert(ann.DatasetID(), ann.ID, &model.AnnotationConvert{
		From: model.AnnotationFormatYolo,
		To:   model.AnnotationFormatAlto,
	})
	if err != nil {
		return "", fmt.Errorf("failed to convert YOLO to ALTO for escriptorium upload: %w", err)
	}
	return ann.AltoDir, nil
}
