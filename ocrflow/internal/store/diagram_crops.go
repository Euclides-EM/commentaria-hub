package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
)

type DiagramCropsStore struct {
	fileSysMgt      *filesys.Manager
	diagramsURLBase string
}

func NewDiagramCropsStore(fileSysMgt *filesys.Manager, diagramsURLBase string) *DiagramCropsStore {
	return &DiagramCropsStore{
		fileSysMgt:      fileSysMgt,
		diagramsURLBase: diagramsURLBase,
	}
}

func (s *DiagramCropsStore) readEditionDiagramsFile(key string) (*editionDiagramsFileData, error) {
	path := s.fileSysMgt.DiagramCropsMetadataFile(key)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("fail to read diagram crops metadata for %s: %w", key, err)
	}

	var fileData editionDiagramsFileData
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("decode diagrams for %s: %w", key, err)
	}
	return &fileData, nil
}

func (s *DiagramCropsStore) GetEditionDiagramCropRelPaths(key string) ([]string, error) {
	fileData, err := s.readEditionDiagramsFile(key)
	if err != nil || fileData == nil {
		return nil, err
	}
	var relPaths []string
	if len(fileData.Volumes) > 0 {
		for _, volume := range fileData.Volumes {
			volumeKey := volume.Key
			if volumeKey == "" {
				volumeKey = key
			}
			for _, imageName := range volume.Images {
				relPaths = append(relPaths, path.Join(volumeKey, "crops", imageName))
			}
		}
		return relPaths, nil
	}
	singleKey := fileData.Key
	if singleKey == "" {
		singleKey = key
	}
	for _, imageName := range fileData.Images {
		relPaths = append(relPaths, path.Join(singleKey, "crops", imageName))
	}
	return relPaths, nil
}

func (s *DiagramCropsStore) GetEditionDiagrams(key string) (*model.DiagramCrops, error) {
	fileData, err := s.readEditionDiagramsFile(key)
	if err != nil {
		return nil, err
	}
	if fileData == nil {
		return &model.DiagramCrops{
			ImageURLsByName: map[string]string{},
			HasDiagrams:     false,
		}, nil
	}
	response := &model.DiagramCrops{
		Key:             fileData.Key,
		HasDiagrams:     fileData.HasDiagrams,
		ImageURLsByName: map[string]string{},
	}

	if len(fileData.Volumes) > 0 {
		response.Volumes = make([]*model.DiagramCropVolume, 0, len(fileData.Volumes))
		for i := range fileData.Volumes {
			volumeKey := fileData.Volumes[i].Key
			if volumeKey == "" {
				volumeKey = key
			}
			response.Volumes = append(response.Volumes, &model.DiagramCropVolume{
				Volume:      fileData.Volumes[i].Volume,
				Key:         fileData.Volumes[i].Key,
				HasDiagrams: fileData.Volumes[i].HasDiagrams,
				ImageURLsByName: mapDiagramImageURLsByName(
					s.diagramsURLBase,
					volumeKey,
					fileData.Volumes[i].Images,
				),
			})
		}
		return response, nil
	}

	singleKey := fileData.Key
	if singleKey == "" {
		singleKey = key
	}
	response.ImageURLsByName = mapDiagramImageURLsByName(
		s.diagramsURLBase,
		singleKey,
		fileData.Images,
	)
	return response, nil
}

func mapDiagramImageURLsByName(diagramsURLBase, key string, images []string) map[string]string {
	out := make(map[string]string, len(images))
	for _, imageName := range images {
		out[imageName] = buildDiagramImageURL(diagramsURLBase, key, imageName)
	}
	return out
}

func buildDiagramImageURL(diagramsURLBase, key, imageName string) string {
	if diagramsURLBase != "" {
		base, err := url.Parse(diagramsURLBase)
		if err == nil {
			base.Path = fmt.Sprintf(
				"%s/%s/crops/%s",
				strings.TrimRight(base.Path, "/"),
				url.PathEscape(key),
				url.PathEscape(imageName),
			)
			return base.String()
		}
	}
	return ""
}

type editionDiagramFileVolume struct {
	Volume      int      `json:"volume,omitempty"`
	Key         string   `json:"key,omitempty"`
	Images      []string `json:"images"`
	HasDiagrams bool     `json:"hasDiagrams"`
}
type editionDiagramsFileData struct {
	Key         string                     `json:"key,omitempty"`
	Images      []string                   `json:"images,omitempty"`
	HasDiagrams bool                       `json:"hasDiagrams"`
	Volumes     []editionDiagramFileVolume `json:"volumes,omitempty"`
}
