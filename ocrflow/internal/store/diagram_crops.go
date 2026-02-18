package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
)

const (
	diagramsPathBaseInGithubFacsimileRepo = "diagrams"
	diagramsCropsDirInGithubFacsimileRepo = "crops"
)

type DiagramCropsStore struct {
	fileSysMgt              *filesys.Manager
	facsimilesGithubRepoUrl string
}

func NewDiagramCropsStore(fileSysMgt *filesys.Manager, facsimilesGithubRepoUrl string) *DiagramCropsStore {
	return &DiagramCropsStore{
		fileSysMgt:              fileSysMgt,
		facsimilesGithubRepoUrl: facsimilesGithubRepoUrl,
	}
}

func (s *DiagramCropsStore) GetEditionDiagrams(key string) (*model.DiagramCrops, error) {
	path := s.fileSysMgt.DiagramCropsMetadataFile(key)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &model.DiagramCrops{
				ImageURLsByName: map[string]string{},
				HasDiagrams:     false,
			}, nil
		}
		return nil, fmt.Errorf("fail to read diagram crops metadata for %s: %w", key, err)
	}

	var fileData editionDiagramsFileData
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("decode diagrams for %s: %w", key, err)
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
					s.facsimilesGithubRepoUrl,
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
		s.facsimilesGithubRepoUrl,
		singleKey,
		fileData.Images,
	)
	return response, nil
}

func mapDiagramImageURLsByName(baseURL, key string, images []string) map[string]string {
	out := make(map[string]string, len(images))
	for _, imageName := range images {
		out[imageName] = buildDiagramImageURL(baseURL, key, imageName)
	}
	return out
}

func buildDiagramImageURL(baseURL, key, imageName string) string {
	return fmt.Sprintf(
		"%s/raw/main/docs/%s/%s/%s/%s",
		baseURL,
		diagramsPathBaseInGithubFacsimileRepo,
		url.PathEscape(key),
		diagramsCropsDirInGithubFacsimileRepo,
		url.PathEscape(imageName),
	)
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
