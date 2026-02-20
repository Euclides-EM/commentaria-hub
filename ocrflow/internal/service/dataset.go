package service

import (
	"context"
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/querylang"
	"github.com/samber/lo"
)

type Dataset struct {
	editionSvc       *Edition
	facsimileSvc     *Facsimile
	modelSvc         *Model
	datasetStore     *store.DatasetSQL
	fileSysMgt       *filesys.Manager
	githubDownloader *ghwrapper.Wrapper
}

func NewDatasetService(editionSvc *Edition, facsimileSvc *Facsimile, modelSvc *Model, datasetStore *store.DatasetSQL, fileSystemMgt *filesys.Manager, githubDownloader *ghwrapper.Wrapper) *Dataset {
	return &Dataset{
		editionSvc:       editionSvc,
		facsimileSvc:     facsimileSvc,
		modelSvc:         modelSvc,
		datasetStore:     datasetStore,
		fileSysMgt:       fileSystemMgt,
		githubDownloader: githubDownloader,
	}
}

func (d *Dataset) List(filter *querylang.Filter, sort querylang.Sort) ([]*model.Dataset, error) {
	// todo: add filtering and sorting and make sure to use it in internal uses in this function
	return d.datasetStore.ListDatasets()
}

func (d *Dataset) Get(id string) (*model.Dataset, error) {
	ds, err := d.datasetStore.GetDataset(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset from store: %w", err)
	}
	if ds == nil {
		return nil, fmt.Errorf("dataset with id %s not found", id)
	}
	return ds, nil
}

func (d *Dataset) Create(ctx context.Context, ds *model.Dataset, enforceSingleDataset, async bool) (*model.Dataset, error) {
	if ds.FacsimileID == "" {
		return nil, fmt.Errorf("currently only datasets linked to facsimiles are supported")
	}
	targetFacsimile, err := d.facsimileSvc.GetFacsimile(ds.FacsimileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}
	if targetFacsimile.ScanURL == "" {
		return nil, fmt.Errorf("facsimile has no scan URL")
	}
	if !ghwrapper.IsGitHubTreeURL(targetFacsimile.ScanURL) {
		return nil, fmt.Errorf("only GitHub tree URLs are supported currently")
	}
	if ds.EditionID != "" && targetFacsimile.EditionID != ds.EditionID {
		return nil, fmt.Errorf("dataset edition ID %s does not match facsimile edition ID %s", ds.EditionID, targetFacsimile.EditionID)
	}
	if ds.EditionID == "" {
		ds.EditionID = targetFacsimile.EditionID
	}

	if enforceSingleDataset {
		dss, err := d.List(nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list datasets: %w", err)
		}
		for _, existingDS := range dss {
			if existingDS.FacsimileID == ds.FacsimileID && existingDS.EditionID == ds.EditionID {
				return nil, fmt.Errorf("dataset for facsimile %s in edition %s already exists", ds.FacsimileID, ds.EditionID)
			}
		}
	}

	ds.ID = idgen.GenerateID(store.DatasetIDPrefix)
	ds.CreatedAt = time.Now()
	ds.UpdatedAt = ds.CreatedAt

	if ds.DPI == 0 || ds.DPI < 50.0 || ds.DPI > 600.0 {
		ds.DPI = 300.0
	}

	if async {
		ds.Status = model.DatasetStatusCreating
		if err := d.datasetStore.InsertDataset(ds); err != nil {
			return nil, fmt.Errorf("failed to insert dataset into store: %w", err)
		}
		// Run heavy work in background; copy fields needed in goroutine to avoid races.
		dsCopy := *ds
		scanURL := targetFacsimile.ScanURL
		go d.runDatasetCreation(context.Background(), &dsCopy, scanURL)
		return ds, nil
	}

	return d.doDatasetCreation(ctx, ds, targetFacsimile.ScanURL)
}

// runDatasetCreation performs download, PDF→PNG, and optional deskew, then updates dataset status.
func (d *Dataset) runDatasetCreation(ctx context.Context, ds *model.Dataset, scanURL string) {
	_, err := d.doDatasetCreation(ctx, ds, scanURL)
	if err != nil {
		log.Printf("async dataset creation failed for %s: %v", ds.ID, err)
		_ = d.datasetStore.UpdateDatasetCreationStatus(ds.ID, model.DatasetStatusFailed, err.Error())
		return
	}
	if err := d.datasetStore.UpdateDatasetCreationStatus(ds.ID, model.DatasetStatusReady, ""); err != nil {
		log.Printf("failed to set dataset %s status to ready: %v", ds.ID, err)
	}
	log.Printf("Dataset %s async creation completed", ds.ID)
}

// doDatasetCreation does download, convert, deskew, and insert (caller sets status when async).
func (d *Dataset) doDatasetCreation(ctx context.Context, ds *model.Dataset, scanURL string) (*model.Dataset, error) {
	pdfPath := d.fileSysMgt.DatasetPDFPath(ds)
	log.Printf("Downloading facsimile from %s to %s", scanURL, pdfPath)
	if err := d.githubDownloader.DownloadRecursive(ctx, scanURL, pdfPath); err != nil {
		return nil, fmt.Errorf("failed to download facsimile: %w", err)
	}

	imgPath := d.fileSysMgt.DatasetImagesDir(ds)
	convertedPNGsDir := imgPath
	if ds.Deskewed {
		var err error
		convertedPNGsDir, err = os.MkdirTemp("", "ocrflow-dataset-rawimgs-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp dir for raw images: %w", err)
		}
		defer os.RemoveAll(convertedPNGsDir)
	}

	var pageList []int
	if strings.TrimSpace(ds.Pages) != "" {
		var err error
		pageList, err = pagesparser.Parse(ds.Pages)
		if err != nil {
			return nil, fmt.Errorf("invalid pages %q: %w", ds.Pages, err)
		}
	}
	if len(pageList) > 0 {
		log.Printf("Converting facsimile PDF %s to PNGs (pages %s) in %s", pdfPath, ds.Pages, convertedPNGsDir)
		if err := formatcov.PDF2PNGsWithPages(pdfPath, convertedPNGsDir, ds.DPI, pageList); err != nil {
			return nil, fmt.Errorf("failed to pre-process facsimile: %w", err)
		}
	} else {
		log.Printf("Converting facsimile PDF %s to PNGs in %s", pdfPath, convertedPNGsDir)
		if err := formatcov.PDF2PNGs(pdfPath, convertedPNGsDir, ds.DPI); err != nil {
			return nil, fmt.Errorf("failed to pre-process facsimile: %w", err)
		}
	}

	if ds.Deskewed {
		log.Printf("Deskewing images from %s into %s", convertedPNGsDir, imgPath)
		if err := formatcov.DeskewPNGs(convertedPNGsDir, imgPath); err != nil {
			return nil, fmt.Errorf("failed to deskew images: %w", err)
		}
	}

	log.Printf("Dataset %s fully created", ds.ID)
	// When async, record was already inserted with status "creating"; we only update status in runDatasetCreation.
	if ds.Status != model.DatasetStatusCreating {
		if err := d.datasetStore.InsertDataset(ds); err != nil {
			return nil, fmt.Errorf("failed to insert dataset into store: %w", err)
		}
	}
	return ds, nil
}

func (d *Dataset) ListSuggestedAnnotationRules(id string) ([][]annotationrule.AnnotationRule, error) {
	ds, err := d.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	fac, err := d.facsimileSvc.GetFacsimile(ds.FacsimileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}
	if fac == nil {
		return nil, fmt.Errorf("facsimile with id %s not found in edition %s", ds.FacsimileID, ds.EditionID)
	}

	allModels, err := d.modelSvc.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	suggestedSegModel := lo.MaxBy(allModels, func(m1, m2 *model.Model) bool {
		if m2.Type != common.OCRModelTypeSegment {
			return true
		}
		if m1.Type != common.OCRModelTypeSegment {
			return false
		}
		if len(m1.BaseAnnotations) > len(m2.BaseAnnotations) {
			return true
		}
		if len(m1.BaseAnnotations) < len(m2.BaseAnnotations) {
			return false
		}
		return m1.CreatedAt.After(m2.CreatedAt)
	})

	//categoriesToRemove := []string{"MainZone-P--Italics", "MainZone-P--Enunciation", "MainZone-P"}
	categoriesForOverlapRemove := []string{
		"CatchWord",
		"DigitizationArtefactZone",
		"DropCapitalZone",
		"DropCapitalZone-Plain",
		"GraphicZone-Decoration",
		"GraphicZone-Diagram",
		"GraphicZone-Table",
		"MainZone",
		"MainZone-Head--Book",
		"MainZone-Head--Section",
		"MarginTextZone-RomanNumerals",
		"NumberingZone",
		"QuireMarksZone",
		"RunningTitleZone",
	}
	categoriesToExcludeFromLineDetection := []string{
		"CatchWord",
		"DigitizationArtefactZone",
		//"DropCapitalZone",
		//""DropCapitalZone-Plain",
		"GraphicZone-Decoration",
		"GraphicZone-Diagram",
		"GraphicZone-Table",
		"NumberingZone",
		"QuireMarksZone",
		"RunningTitleZone",
	}

	return [][]annotationrule.AnnotationRule{
		{
			annotationrule.NewSlicePagesFixed(fac.MainTextPages),
			annotationrule.NewSegment(suggestedSegModel.ID),
			//annotationrule.NewRemoveCategories(categoriesToRemove),
			annotationrule.NewRemoveOverlap(categoriesForOverlapRemove, 1000),
			annotationrule.NewResolveOverlapWithPriority("RunningTitleZone", "MainZone-Head--Section", 0.8),
			annotationrule.NewLinesDetect([]string{"MainZone", "MarginTextZone"}, categoriesToExcludeFromLineDetection),
			annotationrule.NewReassignTextLinesByTolerance("MainZone", "MainZone-Head--Book", 5, 0.6),
			annotationrule.NewReassignTextLinesByTolerance("MainZone", "MainZone-Head--Section", 5, 0.85),
		},
	}, nil
}

func (d *Dataset) GetPageImage(datasetID string, page int) ([]byte, error) {
	ds, err := d.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	imgPath := d.fileSysMgt.DatasetImagesDir(ds)
	filename := pagesparser.PageToPNGFilename(page)
	if _, err := os.Stat(path.Join(imgPath, filename)); err != nil {
		return nil, fmt.Errorf("no such file %s in existing dataset", filename)
	}
	data, err := os.ReadFile(path.Join(imgPath, filename))
	if err != nil {
		return nil, fmt.Errorf("failed to read page image file: %w", err)
	}
	return data, nil
}

func (d *Dataset) ListSuggestedAnnotationReview(id string) ([][]*annotation.ExpectedBlocks, error) {
	return [][]*annotation.ExpectedBlocks{
		{
			{
				Category: "MainZone-Head--Book",
				SanityChecks: []annotation.ExpectedBlocksSanityType{
					annotation.ExpectedBlocksSanityTypeExact,
				},
				ExpectedBlocks: [][]string{
					{"D. HENRION.", "AV LECTEVR."},
					{"ELEMENT", "PREMIER."},
					{"ELEMENT", "PREMIER."},
					{"ELEMENT", "SECOND."},
					{"ELEMENT", "TROISIESME."},
					{"ELEMENT", "QVATRIESME."},
					{"ELEMENT", "CINQVIESME."},
					{"ELEMENT", "SIXIESME."},
					{"ELEMENT", "SEPTIESME."},
					{"ELEMENT", "HVITIESME."},
					{"ELEMENT", "NEVFIESME."},
					{"ELEMENT", "DIXIESME."},
					{"ELEMENT", "VNZIESME."},
					{"ELEMENT", "DOVZIESME."},
					{"ELEMENT", "TREIZIESME."},
					{"ELEMENT", "QVATORZIESME."},
					{"ELEMENT", "QVINZIESME."},
				},
			},
		}, nil,
	}, nil
}

func (d *Dataset) Delete(id string) error {
	ds, err := d.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get dataset: %w", err)
	}
	if err := d.datasetStore.DeleteDataset(id); err != nil {
		return fmt.Errorf("failed to delete dataset from store: %w", err)
	}
	if err := d.fileSysMgt.DeleteDatasetFiles(ds); err != nil {
		return fmt.Errorf("failed to delete dataset files: %w", err)
	}
	return nil
}

func (d *Dataset) Update(id string, m *model.Dataset) (*model.Dataset, error) {
	existingDS, err := d.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	existingDS.Name = m.Name
	existingDS.Description = m.Description
	existingDS.Pages = m.Pages
	if err := d.datasetStore.UpdateDataset(existingDS); err != nil {
		return nil, fmt.Errorf("failed to update dataset in store: %w", err)
	}
	return existingDS, nil
}

func (d *Dataset) UploadImage(file multipart.File, header *multipart.FileHeader, datasetId string) (*model.ImageUpload, error) {
	dataset, err := d.Get(datasetId)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	ext := strings.TrimPrefix(strings.ToLower(mime.TypeByExtension(header.Filename)), "image/")
	if ext == "" {
		return nil, fmt.Errorf("unable to determine file extension for uploaded image")
	}
	if ext != "png" && ext != "jpeg" && ext != "jpg" {
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}
	return d.datasetStore.UploadImage(dataset, header.Filename, file)
}
