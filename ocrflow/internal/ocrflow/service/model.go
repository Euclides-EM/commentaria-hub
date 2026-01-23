package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/tiendc/go-deepcopy"
)

type Model struct {
	modelStore *store.ModelSQL
	fileSysMgt *filesys.Manager
}

func NewModelService(modelStore *store.ModelSQL, fileSysMgt *filesys.Manager) *Model {
	return &Model{
		modelStore: modelStore,
		fileSysMgt: fileSysMgt,
	}
}

func (m *Model) List() ([]*model.Model, error) {
	modelsList, err := m.modelStore.ListModels()
	if err != nil {
		return nil, fmt.Errorf("failed to list models from store: %w", err)
	}
	return modelsList, nil
}

func (m *Model) Get(id string) (*model.Model, error) {
	retrieved, err := m.modelStore.GetModelByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get model from store: %w", err)
	}
	if retrieved == nil {
		return nil, errors.New("model not found")
	}
	return retrieved, nil
}

func (m *Model) Create(mo *model.Model, modelPath string) error {
	var n *model.Model
	if err := deepcopy.Copy(&n, &mo); err != nil {
		return fmt.Errorf("failed to copy annotation: %w", err)
	}
	n.ID = idgen.GenerateID(store.ModelIDPrefix)
	n.Location = model.OCRModelLocationLocal
	n.LocalPath = fmt.Sprintf("%s.pt", n.ID)
	if err := futils.CopyFile(modelPath, m.fileSysMgt.ModelPath(n)); err != nil {
		return fmt.Errorf("failed to copy model from %s to %s: %w", modelPath, m.fileSysMgt.ModelPath(n), err)
	}
	if err := m.modelStore.InsertModel(n); err != nil {
		return fmt.Errorf("failed to upsert model to store: %w", err)
	}
	return nil
}

func (m *Model) Upload(file multipart.File, filename, name, description string, baseAnnotations []*model.AnnotationReference, baseModelID string) (*model.Model, error) {
	if filename == "" {
		return nil, fmt.Errorf("empty filename")
	}
	ext := filepath.Ext(filename)
	modelType := model.OCRModelTypeFromExt(ext)
	if modelType == model.OCRModelTypeUnknown {
		return nil, fmt.Errorf("unsupported model file extension: %s", ext)
	}

	if name == "" {
		name = strings.TrimSuffix(filename, ext)
	}
	id := idgen.Name2ID(store.ModelIDPrefix, name)
	dstFilename := fmt.Sprintf("%s%s", id, ext)

	mo := &model.Model{
		Meta:            model.NewMeta(id).WithName(name).WithDescription(description),
		Type:            modelType,
		Location:        model.OCRModelLocationLocal,
		LocalPath:       dstFilename,
		BaseAnnotations: baseAnnotations,
		BaseModelID:     baseModelID,
	}

	p := m.fileSysMgt.ModelPath(mo)
	if err := futils.WriteMultipartFileToPath(file, p); err != nil {
		return nil, fmt.Errorf("failed to write multipart file to path: %w", err)
	}

	if err := m.modelStore.InsertModel(mo); err != nil {
		defer os.ReadFile(m.fileSysMgt.ModelPath(mo))
		return nil, fmt.Errorf("failed to upsert model to store: %w", err)
	}

	createdModel, err := m.Get(mo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created model: %w", err)
	}

	return createdModel, nil
}

func (m *Model) Delete(id string, fsClean bool) error {
	mo, err := m.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}
	if err := m.modelStore.DeleteModel(id); err != nil {
		return fmt.Errorf("failed to delete model from store: %w", err)
	}
	if fsClean && mo.LocalPath != "" {
		if err := os.Remove(m.fileSysMgt.ModelPath(mo)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete model file from filesystem: %w", err)
		}
	}
	return nil
}
