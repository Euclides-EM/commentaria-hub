package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/querylang"
	"github.com/samber/lo"
	"log"
)

type Dataset struct {
	m                map[string]*model.Dataset
	githubDownloader *ghwrapper.Downloader
	editionSvc       *Edition
	datasetsDir      string
}

func NewDatasetService(githubDownloader *ghwrapper.Downloader, editionSvc *Edition, datasetsDir string) *Dataset {
	return &Dataset{
		m: map[string]*model.Dataset{
			"68kywp": {
				Meta:       model.NewMeta("68kywp"),
				Facsimile:  model.Reference{ID: "1"},
				Edition:    model.Reference{ID: "Paris_1615"},
				PDFPath:    "data/68kywp/Paris_1615_1.pdf",
				ImagesPath: "data/68kywp/imgs",
			},
		},
		githubDownloader: githubDownloader,
		editionSvc:       editionSvc,
		datasetsDir:      datasetsDir,
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

func (d *Dataset) Create(ds *model.Dataset, forceOverwrite bool) (*model.Dataset, error) {
	if ds.FacsimileID() == "" || ds.EditionID() == "" {
		return nil, fmt.Errorf("currently only datasets linked to facsimiles are supported")
	}
	_, targetFacsimile, err := d.editionSvc.GetFacsimile(ds.EditionID(), ds.FacsimileID())
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}
	if targetFacsimile.ScanURL == "" {
		return nil, fmt.Errorf("facsimile has no scan URL")
	}
	if !ghwrapper.IsGitHubTreeURL(targetFacsimile.ScanURL) {
		return nil, fmt.Errorf("only GitHub tree URLs are supported currently")
	}

	if forceOverwrite {
		dss, err := d.List(nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list datasets: %w", err)
		}
		for _, existingDS := range dss {
			if existingDS.FacsimileID() == ds.FacsimileID() && existingDS.EditionID() == ds.EditionID() {
				return nil, fmt.Errorf("dataset for facsimile %s in edition %s already exists", ds.FacsimileID(), ds.EditionID())
			}
		}
	}
	ds.ID = idgen.GenerateID()
	ds.PDFPath = store.DatasetPDFPath(ds, d.datasetsDir)
	ds.ImagesPath = store.DatasetImagesDir(ds, d.datasetsDir)

	log.Printf("Downloading facsimile from %s to %s", targetFacsimile.ScanURL, ds.PDFPath)
	if err := d.githubDownloader.DownloadRecursive(targetFacsimile.ScanURL, ds.PDFPath); err != nil {
		return nil, fmt.Errorf("failed to download facsimile: %w", err)
	}

	log.Printf("Converting facsimile PDF %s to JPGs in %s", ds.PDFPath, ds.ImagesPath)
	if err := formatcov.PDF2JPGs(ds.PDFPath, ds.ImagesPath); err != nil {
		return nil, fmt.Errorf("failed to pre-process facsimile: %w", err)
	}

	log.Printf("Dataset %s fully created", ds.ID)
	d.m[ds.ID] = ds
	return ds, nil
}
