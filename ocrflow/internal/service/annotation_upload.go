package service

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/client"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/escriptorium"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/formatcov"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/roboflow"
	"github.com/samber/lo"
	"github.com/tiendc/go-deepcopy"
)

type AnnotationsUploader struct {
	annotationSvc        *Annotation
	datasetSvc           *Dataset
	fileSysMgt           *filesys.Manager
	roboflowAPIKey       string
	pythonExecutable     string
	escriptoriumPassword string
	escriptoriumUsername string
	escriptoriumBasePath string
	commentariaAPIKey    string
	commentariaBasePath  string
}

func NewAnnotationsUploader(
	annotationSvc *Annotation,
	datasetSvc *Dataset,
	fileSystemMgt *filesys.Manager,
	roboflowAPIKey string,
	pythonExecutable string,
	escriptoriumUsername string,
	escriptoriumPassword string,
	escriptoriumBasePath string,
	commentariaAPIKey string,
	commentariaBasePath string,
) *AnnotationsUploader {
	return &AnnotationsUploader{
		annotationSvc:        annotationSvc,
		datasetSvc:           datasetSvc,
		fileSysMgt:           fileSystemMgt,
		roboflowAPIKey:       roboflowAPIKey,
		pythonExecutable:     pythonExecutable,
		escriptoriumPassword: escriptoriumPassword,
		escriptoriumUsername: escriptoriumUsername,
		escriptoriumBasePath: escriptoriumBasePath,
		commentariaAPIKey:    commentariaAPIKey,
		commentariaBasePath:  commentariaBasePath,
	}
}

// ensureYoloDirForUpload ensures the annotation has a YOLO directory (converting from ALTO if needed).
// Returns the annotation to use for building the upload path.
func (a *AnnotationsUploader) ensureYoloDirForUpload(ann *annotation.Annotation, datasetID string, id string) (*annotation.Annotation, error) {
	if _, err := os.Stat(a.fileSysMgt.DatasetAnnotationYoloDir(ann)); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to stat YOLO annotations dir for roboflow upload: %w", err)
		}
		converted, err := a.convertAlto2Yolo(datasetID, id)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ALTO to YOLO for roboflow upload: %w", err)
		}
		return converted, nil
	}
	return ann, nil
}

func (a *AnnotationsUploader) doRoboflowUpload(ann *annotation.Annotation, rbu *annotation.UploadRoboflow) error {
	params := roboflow.NewUploadDatasetParams().
		SetAPIKey(lo.Ternary(rbu.APIKey == "", a.roboflowAPIKey, rbu.APIKey)).
		SetWorkspaceID(rbu.WorkspaceID).
		SetDatasetPath(a.fileSysMgt.DatasetAnnotationYoloDir(ann)).
		SetProjectID(rbu.ProjectID).
		SetIsNotGroundTruth(rbu.IsNotGroundTruth)
	return roboflow.UploadDataset(a.pythonExecutable, params)
}

func (a *AnnotationsUploader) UploadToRoboflow(datasetID string, id string, rbu *annotation.UploadRoboflow) (*annotation.Annotation, error) {
	ann, err := a.annotationSvc.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("annotation not found: %w", err)
	}
	if !ann.Segmented {
		return nil, fmt.Errorf("annotation is not segmented, cannot upload to roboflow")
	}
	ann, err = a.ensureYoloDirForUpload(ann, datasetID, id)
	if err != nil {
		return nil, err
	}
	if err = a.doRoboflowUpload(ann, rbu); err != nil {
		return nil, fmt.Errorf("failed to upload to roboflow: %w", err)
	}
	var dst *annotation.Annotation
	if err = deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	return dst, nil
}

// UploadToRoboflowAsync validates the request, returns the annotation immediately,
// and performs the upload in a background goroutine. Use with async=true query param.
func (a *AnnotationsUploader) UploadToRoboflowAsync(datasetID string, id string, rbu *annotation.UploadRoboflow) (*annotation.Annotation, error) {
	ann, err := a.annotationSvc.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("annotation not found: %w", err)
	}
	if !ann.Segmented {
		return nil, fmt.Errorf("annotation is not segmented, cannot upload to roboflow")
	}
	var dst annotation.Annotation
	if err = deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	rbuCopy := *rbu
	go func() {
		annForUpload, err := a.annotationSvc.Get(datasetID, id)
		if err != nil {
			log.Printf("roboflow async upload: get annotation: %v", err)
			return
		}
		annForUpload, err = a.ensureYoloDirForUpload(annForUpload, datasetID, id)
		if err != nil {
			log.Printf("roboflow async upload: %v", err)
			return
		}
		if err = a.doRoboflowUpload(annForUpload, &rbuCopy); err != nil {
			log.Printf("roboflow async upload failed for %s: %v", id, err)
			return
		}
		log.Printf("roboflow async upload completed for annotation %s", id)
	}()
	return &dst, nil
}

func (a *AnnotationsUploader) UploadToEscriptorium(datasetID string, id string, aue *annotation.UploadEscriptorium) (*annotation.Annotation, error) {
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

	if !ann.Segmented {
		return nil, fmt.Errorf("no ALTO annotations found for escriptorium upload")
	}

	pages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages for escriptorium upload: %w", err)
	}
	for _, page := range pages {
		imgPath := path.Join(a.fileSysMgt.DatasetImagesDir(ds), pagesparser.PageToPNGFilename(page))
		if err := client.UploadImage(aue.Document, imgPath); err != nil {
			return nil, fmt.Errorf("failed to upload image to escriptorium for page %d: %w", page, err)
		}
		altoPath := path.Join(a.fileSysMgt.DatasetAnnotationAltoDir(ann), pagesparser.PageToXMLFilename(page))
		if err := client.UploadAnnotation(aue.Document, altoPath); err != nil {
			return nil, fmt.Errorf("failed to upload ALTO to escriptorium for page %d: %w", page, err)
		}
	}
	var dst *annotation.Annotation
	if err := deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	return dst, nil
}

func (a *AnnotationsUploader) UploadToCommentaria(datasetID string, id string, cbu *annotation.UploadCommentaria) (*annotation.Annotation, error) {
	ann, err := a.annotationSvc.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("annotation not found: %w", err)
	}
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset for commentaria upload: %w", err)
	}

	c := client.NewClient(
		lo.Ternary(cbu.APIKey == "", a.commentariaAPIKey, cbu.APIKey),
		lo.Ternary(cbu.BasePath == "", a.commentariaBasePath, cbu.BasePath),
	)

	if err := c.Authenticate(); err != nil {
		return nil, fmt.Errorf("failed to authenticate to commentaria: %w", err)
	}

	facsimilies, err := c.GetFacsimilesByEditionID(ds.EditionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimilies from commentaria: %w", err)
	}
	if len(facsimilies) == 0 {
		return nil, fmt.Errorf("no facsimilies found in commentaria for dataset %s", ds.ID)
	}
	facsimile := facsimilies[0] // we assume there's only one facsimile per dataset, which is the one we want to upload annotations to
	ds.FacsimileID = facsimile.ID

	if cbu.DatasetID == "" {
		log.Printf("creating dataset in commentaria instance %s for dataset %s (%s)", cbu.BasePath, ds.Name, ds.ID)
		remoteDs, err := c.CreateDataset(ds)
		if err != nil {
			return nil, fmt.Errorf("failed to create dataset in commentaria: %w", err)
		}
		log.Printf("created dataset in commentaria instance %s: %v", cbu.BasePath, remoteDs)
		cbu.DatasetID = remoteDs.ID
	}

	upm := &annotation.UploadMetadata{
		DatasetID:          cbu.DatasetID,
		Format:             annotation.FormatAlto,
		Name:               ann.Name,
		Description:        ann.Description,
		Segmented:          ann.Segmented,
		GroundTruth:        ann.GroundTruth,
		Ocred:              ann.Ocred,
		LinesDetected:      ann.LinesDetected,
		Hidden:             ann.Hidden,
		OriginAnnotationID: "",
		OCRModelID:         "",
		SegmentModelID:     "",
	}

	log.Printf("uploading annotation %s (%s) to commentaria instance %s for dataset %s (%s)", ann.Name, ann.ID, cbu.BasePath, ds.Name, ds.ID)
	if _, err := c.UploadAnnotation(cbu.DatasetID, upm, a.fileSysMgt.DatasetAnnotationAltoDir(ann)); err != nil {
		return nil, fmt.Errorf("failed to upload annotation to commentaria: %w", err)
	}
	log.Printf("completed upload of annotation %s (%s) to commentaria instance %s for dataset %s (%s)", ann.Name, ann.ID, cbu.BasePath, ds.Name, ds.ID)

	var dst *annotation.Annotation
	if err := deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	return dst, nil
}

func (a *AnnotationsUploader) convertAlto2Yolo(datasetID string, id string) (*annotation.Annotation, error) {
	ann, err := a.annotationSvc.Get(datasetID, id)
	if err != nil {
		return nil, fmt.Errorf("annotation not found: %w", err)
	}
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	if !ann.Segmented {
		return nil, fmt.Errorf("no ALTO annotations found for conversion")
	}
	if err := os.RemoveAll(a.fileSysMgt.DatasetAnnotationYoloDir(ann)); err != nil {
		return nil, fmt.Errorf("failed to clear YOLO annotations dir: %w", err)
	}
	// we could use other segmonto granularities here, the options are: region, subtype, full
	// I didn't notice any difference in the output for different granularities, so I chose "full"...
	if err := formatcov.Alto2Yolo(a.fileSysMgt.DatasetImagesDir(ds), a.fileSysMgt.DatasetAnnotationAltoDir(ann), a.fileSysMgt.DatasetAnnotationYoloDir(ann), 0, "full"); err != nil {
		return nil, fmt.Errorf("failed to convert annotations: %w", err)
	}
	var dst *annotation.Annotation
	if err := deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	return dst, nil
}
