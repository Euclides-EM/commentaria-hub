package service

import (
	"fmt"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/samber/lo"
)

type DatasetImg struct {
	datasetSvc        *Dataset
	fileSysMgt        *filesys.Manager
	datasetImgStore   *store.DatasetImageStore
	tpsTranscriptions *store.TPSTranscriptions
}

func NewDatasetImg(datasetSvc *Dataset, fileSysMgt *filesys.Manager, datasetImgStore *store.DatasetImageStore, tpsTranscriptions *store.TPSTranscriptions) *DatasetImg {
	return &DatasetImg{
		datasetSvc:        datasetSvc,
		fileSysMgt:        fileSysMgt,
		datasetImgStore:   datasetImgStore,
		tpsTranscriptions: tpsTranscriptions,
	}
}

func (d *DatasetImg) GetPageImage(datasetID string, page int) ([]byte, error) {
	ds, err := d.datasetSvc.Get(datasetID)
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

func (d *DatasetImg) UploadImage(file multipart.File, header *multipart.FileHeader, datasetId string, typ string, key string) (*model.ImageUpload, error) {
	dataset, err := d.datasetSvc.Get(datasetId)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	if !futils.IsImageFile(header.Filename) {
		return nil, fmt.Errorf("unsupported image format: %s", header.Filename)
	}
	ext := filepath.Ext(header.Filename)
	d.fileNameForImage(key, typ, ext)
	return d.datasetImgStore.UploadImage(dataset, header.Filename, file)
}

func (d *DatasetImg) ListImages(datasetId string, uniqueOnly bool) ([]*model.ImageMetadata, error) {
	dataset, err := d.datasetSvc.Get(datasetId)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	images, err := d.datasetImgStore.ListImages(dataset)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	if datasetId != "tps" {
		return images, nil
	}
	tpsImages := make(map[string][]*model.ImageMetadata)
	transcribedTPSKeys, err := d.tpsTranscriptions.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to get TPS transcription keys: %w", err)
	}
	transcribedTPSKeysSet := make(map[string]struct{})
	for _, key := range transcribedTPSKeys {
		transcribedTPSKeysSet[key] = struct{}{}
	}
	for _, img := range images {
		key, ok := d.keyFromImageName(img.Filename, "tp")
		if !ok {
			continue
		}
		if _, transcribed := transcribedTPSKeysSet[key]; !transcribed {
			continue
		}
		existing, ok := tpsImages[key]
		if !uniqueOnly || !ok || len(existing) == 0 {
			img.Key = key
			tpsImages[key] = append(tpsImages[key], img)
		} else if img.ModifiedAt.After(existing[0].ModifiedAt) {
			img.Key = key
			tpsImages[key] = []*model.ImageMetadata{img}
		}
	}
	l := lo.Flatten(lo.Values(tpsImages))
	sort.Slice(l, func(i, j int) bool {
		return strings.Compare(l[i].Filename, l[j].Filename) < 0
	})
	return l, nil
}

func (d *DatasetImg) DeleteImages(datasetId string, pageNumOrKeys, filenames []string) error {
	dataset, err := d.datasetSvc.Get(datasetId)
	if err != nil {
		return fmt.Errorf("failed to get dataset: %w", err)
	}
	images, err := d.ListImages(dataset.ID, false)
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}
	var targetImages []*model.ImageMetadata
	for _, img := range images {
		key, ok := d.keyFromImageName(img.Filename, "tp")
		if ok && slices.Contains(pageNumOrKeys, key) {
			targetImages = append(targetImages, img)
			break
		}
		if lo.Contains(filenames, img.Filename) {
			targetImages = append(targetImages, img)
			break
		}
	}
	return d.datasetImgStore.DeleteImages(dataset, targetImages)
}

func (d *DatasetImg) fileNameForImage(key string, typ string, ext string) string {
	return fmt.Sprintf("%s%s", idgen.Name2ID("", fmt.Sprintf("%s_%s", key, typ)), ext)
}

func (d *DatasetImg) keyFromImageName(filename, typ string) (string, bool) {
	fn := strings.TrimSuffix(filename, path.Ext(filename))
	if strings.HasSuffix(fn, "_"+typ) {
		return strings.TrimSuffix(fn, "_"+typ), true
	}
	if split := strings.Split(fn, "_"+typ+"_"); len(split) == 2 {
		return split[0], true
	}
	return "", false
}
