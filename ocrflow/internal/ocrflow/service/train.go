package service

import (
	"errors"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/krakenwrapper"
	"github.com/tiendc/go-deepcopy"
	"log"
	"path"
)

type Train struct {
	annotationSvc *Annotation
	modelSvc      *Model
	trainingDir   string
	m             map[string]*model.Training
}

func NewTrainService(annotationSvc *Annotation, modelSvc *Model, trainingDir string) *Train {
	return &Train{
		annotationSvc: annotationSvc,
		modelSvc:      modelSvc,
		trainingDir:   trainingDir,
		m:             make(map[string]*model.Training),
	}
}

func (tm *Train) TrainYolo(t *model.Training) (*model.Training, error) {
	var originModelLocalPath string
	if originModel, _ := tm.modelSvc.Get(t.OriginModel.ID); originModel != nil {
		if originModel.LocalPath == "" {
			return nil, errors.New("origin model has no local path")
		}
		originModelLocalPath = originModel.LocalPath
	}

	// todo: support multiple annotation sets merging
	if len(t.AnnotationSets) != 1 {
		return nil, errors.New("exactly one annotation set must be provided for training")
	}

	var datsetPaths []string
	for _, annSetRef := range t.AnnotationSets {
		annSet, err := tm.annotationSvc.Get(annSetRef.DatasetID, annSetRef.ID)
		if err != nil {
			return nil, err
		}
		if annSet.YoloDir == "" {
			// todo: convert alto to yolo here...
			return nil, errors.New("annotation set has no YOLO data")
		}
		datsetPaths = append(datsetPaths, annSet.YoloDir)
	}

	datsetYmlPath, err := futils.FindFileByExtension(datsetPaths[0], ".yml", ".yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to find datset: %w", err)
	}

	t.ID = idgen.GenerateID()
	outputPath := store.TrainingDir(t, tm.trainingDir)
	errCh, err := krakenwrapper.TrainYOLOModel(originModelLocalPath, datsetYmlPath, outputPath)
	if err != nil {
		return nil, err
	}

	t.Status = model.TrainingStatusRunning
	tm.m[t.ID] = t
	var toUpdate *model.Training
	if err := deepcopy.Copy(&toUpdate, &t); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	oPath := outputPath

	go func() {
		if recErr := <-errCh; recErr != nil {
			log.Printf("async annotation failed for %s: %v", toUpdate.ID, recErr)
			toUpdate.Status = model.TrainingStatusFailed
			var dst *model.Training
			err := deepcopy.Copy(&dst, &toUpdate)
			if err != nil {
				log.Printf("ERROR failed to deepcopy training model %s: %v", toUpdate.ID, err)
			}
			tm.m[toUpdate.ID] = dst
			return
		}
		log.Printf("async annotation completed for %s", toUpdate.ID)

		toUpdate.Status = model.TrainingStatusCompleted
		var dst *model.Training
		err = deepcopy.Copy(&dst, &toUpdate)
		tm.m[toUpdate.ID] = dst
		if err != nil {
			log.Printf("ERROR failed to deepcopy training model %s: %v", toUpdate.ID, err)
		}
		m := &model.Model{
			Type:            model.OCRModelTypeSegment,
			AlgorithmFamily: model.OCRModelAlgorithmFamilyYOLO,
			Name:            t.Name,
			//	Categories...
		}
		if err := tm.modelSvc.Upsert(m, path.Join(oPath, "weights", "best.pt")); err != nil {
			log.Printf("failed to upsert trained model %s: %v", toUpdate.ID, err)
			return
		}
	}()

	return nil, nil
}
