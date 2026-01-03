package service

import (
	"fmt"
	"log"
	"os"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/querylang"
	"github.com/samber/lo"
)

type Dataset struct {
	m                map[string]*model.Dataset
	githubDownloader *ghwrapper.Downloader
	editionSvc       *Edition
	dataDir          string
}

func NewDatasetService(githubDownloader *ghwrapper.Downloader, editionSvc *Edition, dataDir string) *Dataset {
	return &Dataset{
		m: map[string]*model.Dataset{
			// No deskewing
			"rrpbnk": {
				Meta:        model.NewMeta("rrpbnk"),
				Description: "Dataset without deskewing applied",
				FacsimileID: "2",
				EditionID:   "Paris_1615",
				PDFPath:     "store/data/rrpbnk/Paris_1615_1.pdf",
				ImagesPath:  "store/data/rrpbnk/imgs",
				DPI:         300.0,
			},
			// After deskewing
			"uk5wbj": {
				Meta:        model.NewMeta("uk5wbj"),
				Description: "Dataset with deskewing applied",
				FacsimileID: "1",
				EditionID:   "Paris_1615",
				PDFPath:     "store/data/uk5wbj/Paris_1615_1.pdf",
				ImagesPath:  "store/data/uk5wbj/imgs",
				DPI:         300.0,
			},
			"aiqcec": {
				Meta:        model.NewMeta("aiqcec"),
				FacsimileID: "1",
				EditionID:   "London_1570",
				PDFPath:     "store/data/aiqcec/London_1570_1.pdf",
				ImagesPath:  "store/data/aiqcec/imgs",
				DPI:         300.0,
			},
			"nu3e82": {
				Meta:        model.NewMeta("nu3e82"),
				Description: "Dataset without deskewing applied",
				FacsimileID: "1",
				EditionID:   "Paris_1598a",
				PDFPath:     "store/data/nu3e82/Paris_1598a_1.pdf",
				ImagesPath:  "store/data/nu3e82/imgs",
				DPI:         300.0,
			},
			"mq9w7q": {
				Meta:        model.NewMeta("mq9w7q"),
				Description: "Dataset with deskewing applied",
				FacsimileID: "2",
				EditionID:   "Paris_1598a",
				PDFPath:     "store/data/mq9w7q/Paris_1598a_2.pdf",
				ImagesPath:  "store/data/mq9w7q/imgs",
				DPI:         300.0,
			},
		},
		githubDownloader: githubDownloader,
		editionSvc:       editionSvc,
		dataDir:          dataDir,
	}
}

func (d *Dataset) List(filter *querylang.Filter, sort querylang.Sort) ([]*model.Dataset, error) {
	// todo: add filtering and sorting and make sure to use it in internal uses in this function
	return lo.Values(d.m), nil
}

func (d *Dataset) Get(id string) (*model.Dataset, error) {
	if ds, ok := d.m[id]; ok {
		return ds, nil
	} else {
		return nil, fmt.Errorf("dataset not found")
	}
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
	ds.PDFPath = store.DatasetPDFPath(ds, d.dataDir)
	ds.ImagesPath = store.DatasetImagesDir(ds, d.dataDir)

	if ds.DPI == 0 || ds.DPI < 50.0 || ds.DPI > 600.0 {
		ds.DPI = 300.0
	}

	log.Printf("Downloading facsimile from %s to %s", targetFacsimile.ScanURL, ds.PDFPath)
	if err := d.githubDownloader.DownloadRecursive(targetFacsimile.ScanURL, ds.PDFPath); err != nil {
		return nil, fmt.Errorf("failed to download facsimile: %w", err)
	}

	convertedPNGsDir := ds.ImagesPath
	if !skipDeskew {
		convertedPNGsDir, err = os.MkdirTemp("", "ocrflow-dataset-rawimgs-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp dir for raw images: %w", err)
		}
		defer os.RemoveAll(convertedPNGsDir)
	}

	log.Printf("Converting facsimile PDF %s to PNGs in %s", ds.PDFPath, convertedPNGsDir)
	if err := formatcov.PDF2PNGs(ds.PDFPath, convertedPNGsDir, ds.DPI); err != nil {
		return nil, fmt.Errorf("failed to pre-process facsimile: %w", err)
	}

	if !skipDeskew {
		log.Printf("Deskewing images from %s into %s", convertedPNGsDir, ds.ImagesPath)
		if err := formatcov.DeskewPNGs(convertedPNGsDir, ds.ImagesPath); err != nil {
			return nil, fmt.Errorf("failed to deskew images: %w", err)
		}
	}

	log.Printf("Dataset %s fully created", ds.ID)
	d.m[ds.ID] = ds
	return ds, nil
}

func (d *Dataset) ListSuggestedAnnotationRules(id string) ([][]annotationrule.AnnotationRule, error) {
	return [][]annotationrule.AnnotationRule{
		{
			annotationrule.NewSegment("1615FineTunedCapricciosaM_0312"),
			// pages 321-387 are algebra summary - impossible to detect... todo: Vincenzo
			annotationrule.NewSlicePagesFixed("15-320,388-655"),
			annotationrule.NewRemoveCategories([]string{"MainZone-P--Italics", "MainZone-P--Enunciation", "MainZone-P"}),
			annotationrule.NewRemoveOverlap([]string{"DigitizationArtefactZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "MainZone", "MainZone-Head--Book", "MainZone-Head--Section", "NumberingZone", "QuireMarksZone", "RunningTitleZone"}, 1000),
			annotationrule.NewLinesDetect([]string{"MainZone"}, []string{"CatchWord", "DigitizationArtefactZone", "DropCapitalZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "NumberingZone", "QuireMarksZone", "RunningTitleZone"}),
		},
		{
			annotationrule.NewSegment("1615FineTunedCapricciosaM_0812"),
			annotationrule.NewSlicePagesFixed("15-320,388-655"),
			annotationrule.NewRemoveCategories([]string{"MainZone-P--Italics", "MainZone-P"}),
			annotationrule.NewRemoveOverlap([]string{"DigitizationArtefactZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "MainZone", "MainZone-Head--Book", "MainZone-Head--Section", "NumberingZone", "QuireMarksZone", "RunningTitleZone", "MainZone-P--Enunciation"}, 1000),
			annotationrule.NewLinesDetect([]string{"MainZone"}, []string{"CatchWord", "DigitizationArtefactZone", "DropCapitalZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "NumberingZone", "QuireMarksZone", "RunningTitleZone"}),
		},
	}, nil
}
