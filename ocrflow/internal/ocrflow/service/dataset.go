package service

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/querylang"
)

type Dataset struct {
	editionSvc       *Edition
	datasetStore     *store.DatasetSQL
	fileSysMgt       *filesys.Manager
	githubDownloader *ghwrapper.Downloader
}

func NewDatasetService(editionSvc *Edition, datasetStore *store.DatasetSQL, fileSystemMgt *filesys.Manager, githubDownloader *ghwrapper.Downloader) *Dataset {
	return &Dataset{
		editionSvc:       editionSvc,
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

func (d *Dataset) Create(ds *model.Dataset, forceOverwrite, skipDeskew bool) (*model.Dataset, error) {
	if ds.FacsimileID == "" || ds.EditionID == "" {
		return nil, fmt.Errorf("currently only datasets linked to facsimiles are supported")
	}
	_, targetFacsimile, err := d.editionSvc.GetFacsimile(ds.EditionID, ds.FacsimileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}
	if targetFacsimile.ScanURL == "" {
		return nil, fmt.Errorf("facsimile has no scan URL")
	}
	if !ghwrapper.IsGitHubTreeURL(targetFacsimile.ScanURL) {
		return nil, fmt.Errorf("only GitHub tree URLs are supported currently")
	}

	if !forceOverwrite {
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
	ds.ID = idgen.GenerateID()

	if ds.DPI == 0 || ds.DPI < 50.0 || ds.DPI > 600.0 {
		ds.DPI = 300.0
	}

	pdfPath := d.fileSysMgt.DatasetPDFPath(ds)
	log.Printf("Downloading facsimile from %s to %s", targetFacsimile.ScanURL, pdfPath)
	if err := d.githubDownloader.DownloadRecursive(targetFacsimile.ScanURL, pdfPath); err != nil {
		return nil, fmt.Errorf("failed to download facsimile: %w", err)
	}

	imgPath := d.fileSysMgt.DatasetImagesDir(ds)
	convertedPNGsDir := imgPath
	if !skipDeskew {
		convertedPNGsDir, err = os.MkdirTemp("", "ocrflow-dataset-rawimgs-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp dir for raw images: %w", err)
		}
		defer os.RemoveAll(convertedPNGsDir)
	}

	log.Printf("Converting facsimile PDF %s to PNGs in %s", pdfPath, convertedPNGsDir)
	if err := formatcov.PDF2PNGs(pdfPath, convertedPNGsDir, ds.DPI); err != nil {
		return nil, fmt.Errorf("failed to pre-process facsimile: %w", err)
	}

	if !skipDeskew {
		log.Printf("Deskewing images from %s into %s", convertedPNGsDir, imgPath)
		if err := formatcov.DeskewPNGs(convertedPNGsDir, imgPath); err != nil {
			return nil, fmt.Errorf("failed to deskew images: %w", err)
		}
	}

	log.Printf("Dataset %s fully created", ds.ID)
	if err := d.datasetStore.InsertDataset(ds); err != nil {
		return nil, fmt.Errorf("failed to insert dataset into store: %w", err)
	}
	return ds, nil
}

func (d *Dataset) ListSuggestedAnnotationRules(id string) ([][]annotationrule.AnnotationRule, error) {
	return [][]annotationrule.AnnotationRule{
		{
			// pages 321-387 are algebra summary - impossible to detect... todo: Vincenzo
			annotationrule.NewSlicePagesFixed("15-320,388-655"),
			annotationrule.NewSegment("1615FineTunedCapricciosaM_0312"),
			annotationrule.NewRemoveCategories([]string{"MainZone-P--Italics", "MainZone-P--Enunciation", "MainZone-P"}),
			annotationrule.NewRemoveOverlap([]string{"DigitizationArtefactZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "MainZone", "MainZone-Head--Book", "MainZone-Head--Section", "NumberingZone", "QuireMarksZone", "RunningTitleZone"}, 1000),
			annotationrule.NewLinesDetect([]string{"MainZone"}, []string{"CatchWord", "DigitizationArtefactZone", "DropCapitalZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "NumberingZone", "QuireMarksZone", "RunningTitleZone"}),
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

func (d *Dataset) ListSuggestedAnnotationReview(id string) ([][]*model.AnnotationExpectedBlocks, error) {
	return [][]*model.AnnotationExpectedBlocks{
		{
			{
				Category: "MainZone-Head--Book",
				SanityChecks: []model.AnnotationExpectedBlocksSanityType{
					model.AnnotationExpectedBlocksSanityTypeExact,
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
