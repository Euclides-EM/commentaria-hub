package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

type Execution struct {
	store *fpstore.FeatureExecutionSQL
	mu    sync.Mutex // Protects status updates in goroutines
}

// NewExecution returns a new Execution service using the given store (e.g. *storefeatureplat.FeatureExecutionSQL).
func NewExecution(store *fpstore.FeatureExecutionSQL) *Execution {
	return &Execution{store: store}
}

func (fe *Execution) ListFeatureExecutions(datasetID string, featureIds []string, statuses []feature.ExecutionStatus) ([]*feature.Execution, error) {
	return fe.store.List(datasetID, featureIds, statuses)
}

func (fe *Execution) GetFeatureExecution(executionId string) (*feature.Execution, error) {
	return fe.store.GetByID(executionId)
}

func (fe *Execution) CreateFeatureExecution(exec *feature.Execution) (*feature.Execution, error) {
	exec.ID = idgen.GenerateID("exec")
	exec.Status = feature.ExecutionStatusInProgress
	if err := fe.store.Create(exec); err != nil {
		return nil, err
	}

	// Start a goroutine that will update the status after 60 seconds
	go func(executionID string) {
		time.Sleep(60 * time.Second)
		fe.mu.Lock()
		defer fe.mu.Unlock()
		// Check current status to avoid overwriting if it was canceled
		currentExec, err := fe.store.GetByID(executionID)
		if err != nil {
			return
		}
		// Only update if still in progress
		if currentExec.Status == feature.ExecutionStatusInProgress {
			// 80% probability of success, 20% probability of failed
			if rand.Float64() < 0.8 {
				_ = fe.store.UpdateStatus(executionID, feature.ExecutionStatusSuccess)
			} else {
				_ = fe.store.UpdateStatus(executionID, feature.ExecutionStatusFailed)
			}
		}
	}(exec.ID)

	return exec, nil
}

func (fe *Execution) CancelFeatureExecution(executionId string) (*feature.Execution, error) {
	exec, err := fe.store.GetByID(executionId)
	if err != nil {
		return nil, err
	}
	if exec.Status != feature.ExecutionStatusInProgress && exec.Status != feature.ExecutionStatusCanceling {
		return nil, fmt.Errorf("feature execution cannot be canceled as it is not in progress or already canceling")
	}
	if err := fe.store.UpdateStatus(executionId, feature.ExecutionStatusCanceling); err != nil {
		return nil, err
	}
	exec.Status = feature.ExecutionStatusCanceling

	// Start a goroutine that will update the status from canceling to canceled after 30 seconds
	go func(executionID string) {
		time.Sleep(30 * time.Second)
		fe.mu.Lock()
		defer fe.mu.Unlock()
		// Check current status to ensure it's still canceling
		currentExec, err := fe.store.GetByID(executionID)
		if err != nil {
			return
		}
		// Only update if still in canceling state
		if currentExec.Status == feature.ExecutionStatusCanceling {
			_ = fe.store.UpdateStatus(executionID, feature.ExecutionStatusCanceled)
		}
	}(executionId)

	return exec, nil
}
