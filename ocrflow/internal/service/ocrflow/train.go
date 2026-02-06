package ocrflow

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/krakenwrapper"
	"github.com/tiendc/go-deepcopy"
)

type Train struct {
	annotationSvc *Annotation
	modelSvc      *Model
	fileSysMgt    *filesys.Manager
	trainingDir   string
	m             map[string]*ocrflow.Training
}

func NewTrainService(annotationSvc *Annotation, modelSvc *Model, fileSystemMgt *filesys.Manager, trainingDir string) *Train {
	return &Train{
		annotationSvc: annotationSvc,
		modelSvc:      modelSvc,
		fileSysMgt:    fileSystemMgt,
		trainingDir:   trainingDir,
		m:             make(map[string]*ocrflow.Training),
	}
}

func (tm *Train) TrainYolo(t *ocrflow.Training) (*ocrflow.Training, error) {
	var originModelLocalPath string
	if originModel, _ := tm.modelSvc.Get(t.OriginModel.ID); originModel != nil {
		if originModel.LocalPath == "" {
			return nil, errors.New("origin model has no local path")
		}
		originModelLocalPath = tm.fileSysMgt.ModelPath(t.OriginModel)
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

		if _, err := os.Stat(tm.fileSysMgt.DatasetAnnotationYoloDir(annSet)); err != nil {
			return nil, fmt.Errorf("annotation set YOLO data not found: %w", err)
		}

		datsetPaths = append(datsetPaths, tm.fileSysMgt.DatasetAnnotationYoloDir(annSet))
	}

	datsetYmlPath, err := futils.FindFileByExtension(datsetPaths[0], ".yml", ".yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to find datset: %w", err)
	}

	t.ID = idgen.GenerateID(store.ModelIDPrefix)
	outputPath := tm.fileSysMgt.TrainingDir(t)
	errCh, err := krakenwrapper.TrainYOLOModel(originModelLocalPath, datsetYmlPath, outputPath)
	if err != nil {
		return nil, err
	}

	t.Status = ocrflow.TrainingStatusRunning
	tm.m[t.ID] = t
	var toUpdate *ocrflow.Training
	if err := deepcopy.Copy(&toUpdate, &t); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	oPath := outputPath

	go func() {
		if recErr := <-errCh; recErr != nil {
			log.Printf("async annotation failed for %s: %v", toUpdate.ID, recErr)
			toUpdate.Status = ocrflow.TrainingStatusFailed
			var dst *ocrflow.Training
			err := deepcopy.Copy(&dst, &toUpdate)
			if err != nil {
				log.Printf("ERROR failed to deepcopy training model %s: %v", toUpdate.ID, err)
			}
			tm.m[toUpdate.ID] = dst
			return
		}
		log.Printf("async annotation completed for %s", toUpdate.ID)

		toUpdate.Status = ocrflow.TrainingStatusCompleted
		var dst *ocrflow.Training
		err = deepcopy.Copy(&dst, &toUpdate)
		tm.m[toUpdate.ID] = dst
		if err != nil {
			log.Printf("ERROR failed to deepcopy training model %s: %v", toUpdate.ID, err)
		}
		m := &ocrflow.Model{
			Type:            common.OCRModelTypeSegment,
			AlgorithmFamily: ocrflow.OCRModelAlgorithmFamilyYOLO,
			Meta:            ocrflow.NewMeta(idgen.GenerateID(store.ModelIDPrefix)).WithName(t.Name).WithDescription(t.Description),
			// todo Categories...
		}
		if err := tm.modelSvc.Create(m, path.Join(oPath, "weights", "best.pt")); err != nil {
			log.Printf("failed to upsert trained model %s: %v", toUpdate.ID, err)
			return
		}
	}()

	return nil, nil
}
