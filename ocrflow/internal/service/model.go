package service

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/tiendc/go-deepcopy"
)

var ModelNotFoundErr = errors.New("model not found")

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

// UploadLocalFile imports a model already present on the local filesystem using
// the same validation and persistence path as a multipart API upload.
func (m *Model) UploadLocalFile(modelPath, name, description string) (*model.Model, error) {
	file, err := os.Open(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open model file %s: %w", modelPath, err)
	}
	defer file.Close()

	created, err := m.Upload(file, filepath.Base(modelPath), name, description, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to import model file %s: %w", modelPath, err)
	}
	return created, nil
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
		return nil, ModelNotFoundErr
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

func (m *Model) calcModelID(filename, name string) string {
	ext := filepath.Ext(filename)
	if name == "" {
		name = strings.TrimSuffix(filename, ext)
	}
	id := idgen.Name2ID(store.ModelIDPrefix, name)
	return id
}

func (m *Model) Upload(file multipart.File, filename, name, description string, baseAnnotations []*annotation.Reference, baseModelID string) (*model.Model, error) {
	if filename == "" {
		return nil, fmt.Errorf("empty filename")
	}
	ext := filepath.Ext(filename)
	modelType := common.OCRModelTypeFromExt(ext)
	if modelType == common.OCRModelTypeUnknown {
		return nil, fmt.Errorf("unsupported model file extension: %s", ext)
	}

	if name == "" {
		name = strings.TrimSuffix(filename, ext)
	}
	id := m.calcModelID(filename, name)
	dstFilename := fmt.Sprintf("%s%s", id, ext)

	mo := &model.Model{
		Meta:            common.NewMeta(id).WithName(name).WithDescription(description),
		Type:            modelType,
		Location:        model.OCRModelLocationLocal,
		LocalPath:       dstFilename,
		BaseAnnotations: baseAnnotations,
		BaseModelID:     baseModelID,
	}
	if modelType == common.OCRModelTypeSegment {
		mo.AlgorithmFamily = model.OCRModelAlgorithmFamilyYOLO
	}

	p := m.fileSysMgt.ModelPath(mo)
	if err := futils.WriteMultipartFileToPath(file, p); err != nil {
		return nil, fmt.Errorf("failed to write multipart file to path: %w", err)
	}

	if err := m.modelStore.InsertModel(mo); err != nil {
		defer os.Remove(m.fileSysMgt.ModelPath(mo))
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

func (m *Model) Update(id string, mo *model.Model) (any, error) {
	existing, err := m.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing model: %w", err)
	}

	existing.Meta.Name = mo.Meta.Name
	existing.Meta.Description = mo.Meta.Description
	existing.BaseAnnotations = mo.BaseAnnotations
	existing.BaseModelID = mo.BaseModelID

	if err := m.modelStore.UpdateModel(existing); err != nil {
		return nil, fmt.Errorf("failed to update model in store: %w", err)
	}

	updatedModel, err := m.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated model: %w", err)
	}

	return updatedModel, nil
}

func (m *Model) InitDefaultModels() error {
	defaultModelPath := m.fileSysMgt.DefaultModelPath()
	defaultModels, err := os.ReadDir(defaultModelPath)
	if err != nil {
		return fmt.Errorf("failed to read default models: %w", err)
	}
	for _, defaultModel := range defaultModels {
		if defaultModel.IsDir() {
			continue
		}
		if common.OCRModelTypeFromExt(defaultModel.Name()) == common.OCRModelTypeUnknown {
			continue
		}
		if _, err := m.Get(m.calcModelID(defaultModel.Name(), "")); err == nil {
			continue
		} else if !errors.Is(err, ModelNotFoundErr) {
			return fmt.Errorf("failed to get DB entry for default model: %w", err)
		}
		modelPath := path.Join(defaultModelPath, defaultModel.Name())
		log.Printf("importing default model %s...", modelPath)
		if _, err := m.UploadLocalFile(modelPath, "", ""); err != nil {
			return fmt.Errorf("import default model: %w", err)
		}
		log.Printf("imported default model %s", modelPath)
	}
	return nil
}
