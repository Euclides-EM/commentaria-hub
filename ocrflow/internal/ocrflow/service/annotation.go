package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/krakenwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/roboflow"
	"github.com/samber/lo"
	"log"
	"math/rand"
	"os"
	"path"
)

// todo: add interfaces to all services

type Annotations struct {
	pythonExecutable     string
	m                    map[string]*model.Annotation
	datasetSvc           *Dataset
	modelSvc             *Model
	roboflowAPIKey       string
	ruleApplier          *AnnotationRuleApplier
	escriptoriumBasePath string
	escriptoriumUsername string
	escriptoriumPassword string
}

func NewAnnotationsService(
	pythonExecutable string,
	datasetSvc *Dataset,
	modelSvc *Model,
	roboflowAPIKey string,
	ruleApplier *AnnotationRuleApplier,
	escriptoriumBasePath string,
	escriptoriumUsername string,
	escriptoriumPassword string,
) *Annotations {
	annotations := map[string]*model.Annotation{}
	manuallyAnnotated := map[string]*model.Annotation{
		// Manually annotated

		// Only one big MainZone, without any subtype like paragraph, enunciation, etc.
		// In Robolflow it's here: https://app.roboflow.com/mia-workplace/paris-1615-withmznosubtypes-tkgii/1
		"gog80x": {
			Meta:    model.NewMeta("gog80x"),
			Pages:   "66,160,197,303,497,20,49,91,95-97,148-149,153,183,195-196,255,257,295-297,315,388,395-397,450-465,495-496,508,596,603,624,256,339,387,595,597,609",
			Dataset: model.Reference{ID: "rrpbnk"},
			YoloDir: "store/data/annotations/gog80x/yolo",
		},

		// Polygons for subtypes like paragraph, enunciation and main zones.
		// In Roboflow it's here: https://app.roboflow.com/mia-workplace/paris-1615-polygonswithmz-wsrge/1
		"f0k3ks": {
			Meta:    model.NewMeta("f0k3ks"),
			Pages:   "66,160,197,303,497,20,49,91,95-97,148-149,153,183,195-196,255,257,295-297,315,388,395-397,450-465,495-496,508,596,603,624,256,339,387,595,597,609",
			Dataset: model.Reference{ID: "rrpbnk"},
			YoloDir: "store/data/annotations/f0k3ks/yolo",
		},

		// Includes a big MainZone, in addition to subtypes like paragraph, enunciation, etc.
		// In some cases, instead of boxes polygons were drawn to better fit the text areas.
		// In Roboflow it's here: https://app.roboflow.com/mia-workplace/paris-1615-polygonswithmz-wsrge/1
		"idim36": {
			Meta:    model.NewMeta("idim36"),
			Pages:   "66,160,197,303,497,20,49,91,95-97,148-149,153,183,195-196,255,257,295-297,315,388,395-397,450-465,495-496,508,596,603,624,256,339,387,595,597,609",
			Dataset: model.Reference{ID: "rrpbnk"},
			YoloDir: "store/data/annotations/idim36/yolo",
		},
	}

	for k, v := range manuallyAnnotated {
		annotations[k] = v
		annotations[k].YoloDir = store.DatasetAnnotationYoloDir(annotations[k], datasetSvc.datasetsDir)
	}

	inferredAnnotations := map[string]*model.Annotation{
		// inferred annotations

		// full annotation using the Paris1615Polygons1 model
		"3j2xr7": {
			Meta:              model.NewMeta("3j2xr7"),
			Pages:             "15-655",
			Dataset:           model.Reference{ID: "rrpbnk"},
			SegmentationModel: model.Reference{ID: "Paris1615Polygons1"},
		},

		// full annotation using the Paris1615NoContinuedPNoMainZone3 model
		"4afrrf": {
			Meta:              model.NewMeta("4afrrf"),
			Pages:             "15-655",
			Dataset:           model.Reference{ID: "rrpbnk"},
			SegmentationModel: model.Reference{ID: "Paris1615NoContinuedPNoMainZone3"},
		},

		// full annotation using the Paris1615NoMainZoneSubtypes model
		"h3y0bj": {
			Meta:              model.NewMeta("h3y0bj"),
			Pages:             "15-655",
			Dataset:           model.Reference{ID: "rrpbnk"},
			SegmentationModel: model.Reference{ID: "Paris1615NoMainZoneSubtypes"},
		},

		// full annotation using the Paris1615PolygonsAndMainZone model
		"xu3fkx": {
			Meta:              model.NewMeta("xu3fkx"),
			Pages:             "15-655",
			Dataset:           model.Reference{ID: "rrpbnk"},
			SegmentationModel: model.Reference{ID: "Paris1615PolygonsAndMainZone"},
		},
	}

	for k, v := range inferredAnnotations {
		annotations[k] = v
		annotations[k].RoboflowDir = store.DatasetAnnotationRoboflowDir(annotations[k], datasetSvc.datasetsDir)
	}

	return &Annotations{
		pythonExecutable:     pythonExecutable,
		m:                    annotations,
		datasetSvc:           datasetSvc,
		modelSvc:             modelSvc,
		ruleApplier:          ruleApplier,
		roboflowAPIKey:       roboflowAPIKey,
		escriptoriumBasePath: escriptoriumBasePath,
		escriptoriumUsername: escriptoriumUsername,
		escriptoriumPassword: escriptoriumPassword,
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

func (a *Annotations) Get(datasetId, id string) (*model.Annotation, error) {
	annotation, ok := a.m[id]
	if !ok || annotation.DatasetID() != datasetId {
		return nil, fmt.Errorf("annotation not found")
	}
	return annotation.DeepCopy(), nil
}

func (a *Annotations) Create(datasetID string, ann *model.Annotation, async bool, randomPages int) (*model.Annotation, error) {
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

	var pages []int
	if ann.Pages == "" && randomPages > 0 {
		allPages, err := store.InferPages(ds.ImagesPath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to count dataset pages: %w", err)
		}
		for i := 0; i < randomPages; i++ {
			pages = append(pages, allPages[rand.Intn(len(allPages))])
		}

	} else {
		pages, err = pagesparser.Parse(ann.Pages)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pages: %w", err)
		}
	}

	var filenames []string
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
	switch segM.Location {
	case model.OCRModelLocationLocal:
		ann.AltoDir = store.DatasetAnnotationAltoDir(ann, a.datasetSvc.datasetsDir)
		return krakenwrapper.Recognize(
			ds.ImagesPath,
			ann.AltoDir,
			segM.LocalPath,
			lo.TernaryF(ocrM == nil, func() string { return "" }, func() string { return ocrM.LocalPath }),
			filenames,
		)
	case model.OCRModelLocationRoboflow:
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
	return nil, fmt.Errorf("unsupported model runtime: %s", segM.Location)
}

func (a *Annotations) Convert(datasetID string, id string, annc *model.AnnotationConvert) (*model.Annotation, error) {
	ann, ok := a.m[id]
	if !ok || ann.DatasetID() != datasetID {
		return nil, fmt.Errorf("annotation not found")
	}

	if annc.From == model.AnnotationFormatAlto && annc.To == model.AnnotationFormatYolo {
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

	if annc.From == model.AnnotationFormatYolo && annc.To == model.AnnotationFormatAlto {
		if ann.YoloDir == "" {
			return nil, fmt.Errorf("no YOLO annotations found for conversion")
		}
		ann.AltoDir = store.DatasetAnnotationAltoDir(ann, a.datasetSvc.datasetsDir)
		if err := os.RemoveAll(ann.AltoDir); err != nil {
			return nil, fmt.Errorf("failed to clear ALTO annotations dir: %w", err)
		}
		if err := formatcov.Yolo2Alto(ann.YoloDir, ann.AltoDir); err != nil {
			return nil, fmt.Errorf("failed to convert annotations: %w", err)
		}
		return ann.DeepCopy(), nil
	}
	return nil, fmt.Errorf("unsupported annotation format for conversion %s -> %s", annc.From, annc.To)
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
		dstPath = store.DatasetAnnotationAltoDir(ann, a.datasetSvc.datasetsDir)
		ann.AltoDir = dstPath
	case model.AnnotationFormatYolo:
		dstPath = store.DatasetAnnotationYoloDir(ann, a.datasetSvc.datasetsDir)
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

func (a *Annotations) ApplyRules(datasetID string, id string, aar *model.AnnotationApplyRules) (*model.Annotation, error) {
	fromDB, ok := a.m[id]
	if !ok || fromDB.DatasetID() != datasetID {
		return nil, fmt.Errorf("annotation not found")
	}

	ann := fromDB.DeepCopy()

	//if ann.RoboflowDir == "" {
	//	return nil, fmt.Errorf("currently only Roboflow annotations are supported for rule application")
	//}

	//ann.YoloDir = ""
	//ann.AltoDir = ""

	if ann.YoloDir != "" && ann.RoboflowDir == "" && ann.AltoDir == "" {
		log.Printf("only YOLO annotations found, converting to ALTO for rule application")
		var err error
		ann, err = a.Convert(ann.DatasetID(), ann.ID, &model.AnnotationConvert{
			From: model.AnnotationFormatYolo,
			To:   model.AnnotationFormatAlto,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to convert YOLO to ALTO for rule application: %w", err)
		}
	}
	if aar.Action == model.AnnotationApplyRulesActionCreateNew {
		ann.ID = idgen.GenerateID()
		if fromDB.RoboflowDir != "" {
			ann.RoboflowDir = store.DatasetAnnotationRoboflowDir(ann, a.datasetSvc.datasetsDir)
			if err := futils.CopyDir(store.DatasetAnnotationRoboflowDir(fromDB, a.datasetSvc.datasetsDir), ann.RoboflowDir); err != nil {
				return nil, fmt.Errorf("failed to copy annotations for new annotation: %w", err)
			}
		}
		if fromDB.YoloDir != "" {
			ann.YoloDir = store.DatasetAnnotationYoloDir(ann, a.datasetSvc.datasetsDir)
			if err := futils.CopyDir(store.DatasetAnnotationYoloDir(fromDB, a.datasetSvc.datasetsDir), ann.YoloDir); err != nil {
				return nil, fmt.Errorf("failed to copy annotations for new annotation: %w", err)
			}
		}
		if fromDB.AltoDir != "" {
			ann.AltoDir = store.DatasetAnnotationAltoDir(ann, a.datasetSvc.datasetsDir)
			if err := futils.CopyDir(store.DatasetAnnotationAltoDir(fromDB, a.datasetSvc.datasetsDir), ann.AltoDir); err != nil {
				return nil, fmt.Errorf("failed to copy annotations for new annotation: %w", err)
			}
		}
		a.m[ann.ID] = ann.DeepCopy()
	}

	// apply rules...
	if err := a.ruleApplier.ApplyRules(ann, aar.Rules); err != nil {
		return nil, fmt.Errorf("failed to apply annotation rules: %w", err)
	}

	a.m[ann.ID] = ann.DeepCopy()

	return a.m[ann.ID], nil
}
