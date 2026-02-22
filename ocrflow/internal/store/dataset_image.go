package store

import (
	"fmt"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/samber/lo"
)

type DatasetImageStore struct {
	FilesysManager *filesys.Manager
}

func NewDatasetImageStore(filesysManager *filesys.Manager) *DatasetImageStore {
	return &DatasetImageStore{
		FilesysManager: filesysManager,
	}
}

func (s *DatasetImageStore) UploadImage(ds *model.Dataset, filename string, file multipart.File) (*model.ImageUpload, error) {
	p := path.Join(s.FilesysManager.DatasetImagesDir(ds), filename)
	if err := futils.WriteMultipartFileToPath(file, p); err != nil {
		return nil, fmt.Errorf("Error saving uploaded image: %v\n", err)
	}
	return &model.ImageUpload{
		Success:  true,
		Filename: filepath.Base(p),
		Path:     filepath.Join(filepath.Dir(p), filepath.Base(p)),
	}, nil
}

func (s *DatasetImageStore) ListImages(dataset *model.Dataset) ([]*model.ImageMetadata, error) {
	imgDir := s.FilesysManager.DatasetImagesDir(dataset)
	files, err := os.ReadDir(imgDir)
	if err != nil {
		return nil, fmt.Errorf("Error listing images: %v\n", err)
	}
	var images []*model.ImageMetadata
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !futils.IsImageFile(file.Name()) {
			continue
		}

		var name string
		page, err := pagesparser.FileNameToPage(file.Name())

		name = lo.IfF(err != nil, func() string {
			return strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		}).ElseF(func() string {
			return fmt.Sprintf("%d", page)
		})
		fi, err := file.Info()
		if err != nil {
			return nil, fmt.Errorf("Error getting file info: %v\n", err)
		}

		images = append(images, &model.ImageMetadata{
			Key:        name,
			Filename:   file.Name(),
			ModifiedAt: fi.ModTime(),
		})
	}
	return images, nil
}

func (s *DatasetImageStore) GetImageMetadata(dataset *model.Dataset, key string) ([]*model.ImageMetadata, error) {
	imgDir := s.FilesysManager.DatasetImagesDir(dataset)
	files, err := os.ReadDir(imgDir)
	if err != nil {
		return nil, fmt.Errorf("Error listing images: %v\n", err)
	}
	var images []*model.ImageMetadata
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !futils.IsImageFile(file.Name()) {
			continue
		}
		if strings.HasPrefix(file.Name(), key) {
			fi, err := file.Info()
			if err != nil {
				return nil, fmt.Errorf("Error getting file info: %v\n", err)
			}
			images = append(images, &model.ImageMetadata{
				Key:        strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
				Filename:   file.Name(),
				ModifiedAt: fi.ModTime(),
			})
		}
	}
	return images, nil
}

func (s *DatasetImageStore) DeleteImages(ds *model.Dataset, images []*model.ImageMetadata) error {
	for _, img := range images {
		p := path.Join(s.FilesysManager.DatasetImagesDir(ds), img.Filename)
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("Error deleting image %s: %v\n", img.Filename, err)
		}
	}
	return nil
}
