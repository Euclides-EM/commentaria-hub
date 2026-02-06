package featureplat

import (
	"fmt"
	"slices"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/samber/lo"
)

type Execution struct {
	m map[string]*featureplat.FeatureExecution
}

func NewExecution() *Execution {
	mockMap := make(map[string]*featureplat.FeatureExecution)
	return &Execution{
		m: mockMap,
	}
}

func (fe *Execution) ListFeatureExecutions(collectionId string, featureIds []string, statuses []featureplat.FeatureExecutionStatus) ([]*featureplat.FeatureExecution, error) {
	var executions []*featureplat.FeatureExecution
	for _, exec := range fe.m {
		appliedFeatures := lo.Map(exec.Apply, func(item featureplat.FeatureExecutionApplyItem, _ int) string {
			return item.Feature
		})
		if exec.Collection == collectionId && len(lo.Intersect(featureIds, appliedFeatures)) > 0 && slices.Contains(statuses, exec.Status) {
			executions = append(executions, exec)
		}
	}
	return executions, nil
}

func (fe *Execution) GetFeatureExecution(executionId string) (*featureplat.FeatureExecution, error) {
	if exec, ok := fe.m[executionId]; ok {
		return exec, nil
	}
	return nil, fmt.Errorf("feature execution not found")
}

func (fe *Execution) CreateFeatureExecution(exec *featureplat.FeatureExecution) (*featureplat.FeatureExecution, error) {
	exec.ID = idgen.GenerateID("exec")
	exec.CreatedAt = time.Now()
	exec.UpdatedAt = time.Now()
	exec.Status = featureplat.FeatureExecutionStatusInProgress
	fe.m[exec.ID] = exec
	return exec, nil
}

func (fe *Execution) CancelFeatureExecution(executionId string) (*featureplat.FeatureExecution, error) {
	exec, ok := fe.m[executionId]
	if !ok {
		return nil, fmt.Errorf("feature execution not found")
	}
	if exec.Status != featureplat.FeatureExecutionStatusInProgress && exec.Status != featureplat.FeatureExecutionStatusCanceling {
		return nil, fmt.Errorf("feature execution cannot be canceled as it is not in progress or already canceling")
	}
	exec.Status = featureplat.FeatureExecutionStatusCanceling
	exec.UpdatedAt = time.Now()
	return exec, nil
}
