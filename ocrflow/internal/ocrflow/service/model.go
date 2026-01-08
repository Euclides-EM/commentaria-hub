package service

import (
	"errors"
	"fmt"

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
	n.ID = idgen.GenerateID()
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
