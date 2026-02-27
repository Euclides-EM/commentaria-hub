package store

import (
	"errors"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/cache"
)

const featureExecutionTTL = 24 * time.Hour

var ErrFeatureExecutionNotFound = errors.New("feature execution not found")

type FeatureExecutionStore struct {
	cache *cache.Cache
}

func NewFeatureExecutionStore(c *cache.Cache) *FeatureExecutionStore {
	return &FeatureExecutionStore{cache: c}
}

func (s *FeatureExecutionStore) List(datasetID string, featureIDs []string, statuses []feature.ExecutionStatus) ([]*feature.Execution, error) {
	if datasetID == "" {
		return nil, errors.New("list executions: missing dataset_id")
	}

	featureSet := map[string]struct{}{}
	for _, fid := range featureIDs {
		featureSet[fid] = struct{}{}
	}
	statusSet := map[feature.ExecutionStatus]struct{}{}
	for _, st := range statuses {
		statusSet[st] = struct{}{}
	}

	_, vals, _, err := s.cache.GetBulk(nil, func(k1, k2 string, v1, v2 any) int {
		e1, _ := v1.(*feature.Execution)
		e2, _ := v2.(*feature.Execution)
		// newest first
		return e2.CreatedAt.Compare(e1.CreatedAt)
	}, 0, 10_000)
	if err != nil {
		return nil, err
	}

	var out []*feature.Execution
	for _, v := range vals {
		exec, ok := v.(*feature.Execution)
		if !ok || exec == nil {
			continue
		}

		if exec.DatasetID != datasetID {
			continue
		}

		if len(statusSet) > 0 {
			if _, ok := statusSet[exec.Status]; !ok {
				continue
			}
		}

		// Filter by "featureIDs" meaning: any apply item targets one of these features.
		if len(featureSet) > 0 {
			matched := false
			for _, a := range exec.Apply {
				if _, ok := featureSet[a.Feature]; ok {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		out = append(out, exec)
	}

	return out, nil
}

func (s *FeatureExecutionStore) GetByID(id string) (*feature.Execution, error) {
	if id == "" {
		return nil, fmt.Errorf("get execution: missing id")
	}
	v, ok := s.cache.Get(id)
	if !ok {
		return nil, ErrFeatureExecutionNotFound
	}
	exec, ok := v.(*feature.Execution)
	if !ok || exec == nil {
		return nil, ErrFeatureExecutionNotFound
	}
	return exec, nil
}

func (s *FeatureExecutionStore) Create(exec *feature.Execution) error {
	if exec == nil {
		return errors.New("create execution: nil execution")
	}
	if exec.ID == "" {
		return errors.New("create execution: missing id")
	}
	if exec.DatasetID == "" {
		return errors.New("create execution: missing dataset_id")
	}
	if len(exec.Apply) == 0 {
		return errors.New("create execution: apply is empty")
	}

	now := time.Now()
	if exec.CreatedAt.IsZero() {
		exec.CreatedAt = now
	}
	exec.UpdatedAt = now

	s.cache.SetWithTTL(exec.ID, exec, featureExecutionTTL)
	return nil
}

func (s *FeatureExecutionStore) UpdateStatus(id string, status feature.ExecutionStatus, statusReason string) error {
	exec, err := s.GetByID(id)
	if err != nil {
		return err
	}

	exec.Status = status
	exec.StatusReason = statusReason
	exec.UpdatedAt = time.Now()

	s.cache.SetWithTTL(exec.ID, exec, featureExecutionTTL)
	return nil
}
