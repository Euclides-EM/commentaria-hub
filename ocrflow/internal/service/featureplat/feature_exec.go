package featureplat

import (
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store/featureplat"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

type Execution struct {
	store *fpstore.FeatureExecutionSQL
}

// NewExecution returns a new Execution service using the given store (e.g. *storefeatureplat.FeatureExecutionSQL).
func NewExecution(store *fpstore.FeatureExecutionSQL) *Execution {
	return &Execution{store: store}
}

func (fe *Execution) ListFeatureExecutions(collectionId string, featureIds []string, statuses []featureplat.FeatureExecutionStatus) ([]*featureplat.FeatureExecution, error) {
	return fe.store.List(collectionId, featureIds, statuses)
}

func (fe *Execution) GetFeatureExecution(executionId string) (*featureplat.FeatureExecution, error) {
	return fe.store.GetByID(executionId)
}

func (fe *Execution) CreateFeatureExecution(exec *featureplat.FeatureExecution) (*featureplat.FeatureExecution, error) {
	exec.ID = idgen.GenerateID("exec")
	exec.Status = featureplat.FeatureExecutionStatusInProgress
	if err := fe.store.Create(exec); err != nil {
		return nil, err
	}
	return exec, nil
}

func (fe *Execution) CancelFeatureExecution(executionId string) (*featureplat.FeatureExecution, error) {
	exec, err := fe.store.GetByID(executionId)
	if err != nil {
		return nil, err
	}
	if exec.Status != featureplat.FeatureExecutionStatusInProgress && exec.Status != featureplat.FeatureExecutionStatusCanceling {
		return nil, fmt.Errorf("feature execution cannot be canceled as it is not in progress or already canceling")
	}
	if err := fe.store.UpdateStatus(executionId, featureplat.FeatureExecutionStatusCanceling); err != nil {
		return nil, err
	}
	exec.Status = featureplat.FeatureExecutionStatusCanceling
	return exec, nil
}
