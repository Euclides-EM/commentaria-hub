package store

import (
	"errors"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/integration"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/cache"
)

const integrationJobTTL = 24 * time.Hour

var ErrIntegrationJobNotFound = errors.New("integration job not found")

type IntegrationJobStore struct {
	cache *cache.Cache
}

func NewIntegrationJobStore(c *cache.Cache) *IntegrationJobStore {
	return &IntegrationJobStore{cache: c}
}

func (s *IntegrationJobStore) Create(job *integration.Job) {
	s.cache.SetWithTTL(job.ID, job, integrationJobTTL)
}

func (s *IntegrationJobStore) Get(id string) (*integration.Job, error) {
	v, ok := s.cache.Get(id)
	if !ok {
		return nil, ErrIntegrationJobNotFound
	}
	job, ok := v.(*integration.Job)
	if !ok {
		return nil, ErrIntegrationJobNotFound
	}
	return job, nil
}

func (s *IntegrationJobStore) Update(job *integration.Job) {
	s.cache.SetWithTTL(job.ID, job, integrationJobTTL)
}

func (s *IntegrationJobStore) ListAll() ([]*integration.Job, error) {
	var jobs []*integration.Job
	_, vals, _, err := s.cache.GetBulk(nil, func(k1 string, k2 string, v1 any, v2 any) int {
		j1, _ := v1.(*integration.Job)
		j2, _ := v2.(*integration.Job)
		return j2.CreatedAt.Compare(j1.CreatedAt)
	}, 0, 1000)
	if err != nil {
		return nil, err
	}
	for _, v := range vals {
		if job, ok := v.(*integration.Job); ok {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}
