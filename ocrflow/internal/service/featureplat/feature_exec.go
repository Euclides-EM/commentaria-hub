package featureplat

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store/featureplat"
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
		if currentExec.Status == featureplat.FeatureExecutionStatusInProgress {
			// 80% probability of success, 20% probability of failed
			if rand.Float64() < 0.8 {
				_ = fe.store.UpdateStatus(executionID, featureplat.FeatureExecutionStatusSuccess)
			} else {
				_ = fe.store.UpdateStatus(executionID, featureplat.FeatureExecutionStatusFailed)
			}
		}
	}(exec.ID)

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
		if currentExec.Status == featureplat.FeatureExecutionStatusCanceling {
			_ = fe.store.UpdateStatus(executionID, featureplat.FeatureExecutionStatusCanceled)
		}
	}(executionId)

	return exec, nil
}
