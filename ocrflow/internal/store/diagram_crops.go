package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

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
	diagramsContentBase := resolveDiagramsContentBase(s.facsimilesGithubRepoUrl)
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
					diagramsContentBase,
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
		diagramsContentBase,
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
	trimmedBase := strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf(
		"%s/%s/%s/%s/%s",
		trimmedBase,
		diagramsPathBaseInGithubFacsimileRepo,
		url.PathEscape(key),
		diagramsCropsDirInGithubFacsimileRepo,
		url.PathEscape(imageName),
	)
}

func resolveDiagramsContentBase(facsimilesRepoURL string) string {
	repoURL := strings.TrimSuffix(strings.TrimSpace(facsimilesRepoURL), "/")
	parsed, err := url.Parse(repoURL)
	if err != nil {
		return repoURL
	}

	if strings.EqualFold(parsed.Host, "raw.githubusercontent.com") {
		return repoURL
	}

	if !strings.EqualFold(parsed.Host, "github.com") {
		return repoURL
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "blob" {
		return repoURL
	}

	contentPath := append([]string{parts[0], parts[1], parts[3]}, parts[4:]...)
	return "https://raw.githubusercontent.com/" + strings.Join(contentPath, "/")
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
