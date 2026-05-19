package store

import (
	"errors"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/job"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/cache"
)

const jobTTL = 24 * time.Hour

var ErrJobNotFound = errors.New("job not found")

type JobStore struct {
	cache *cache.Cache
}

func NewJobStore(c *cache.Cache) *JobStore {
	return &JobStore{cache: c}
}

func (s *JobStore) Create(job *job.Job) {
	s.cache.SetWithTTL(job.ID, job, jobTTL)
}

func (s *JobStore) Get(id string) (*job.Job, error) {
	v, ok := s.cache.Get(id)
	if !ok {
		return nil, ErrJobNotFound
	}
	job, ok := v.(*job.Job)
	if !ok {
		return nil, ErrJobNotFound
	}
	return job, nil
}

func (s *JobStore) Update(job *job.Job) {
	s.cache.SetWithTTL(job.ID, job, jobTTL)
}

func (s *JobStore) ListAll() ([]*job.Job, error) {
	var jobs []*job.Job
	_, vals, _, err := s.cache.GetBulk(nil, func(k1 string, k2 string, v1 any, v2 any) int {
		j1, _ := v1.(*job.Job)
		j2, _ := v2.(*job.Job)
		return j2.CreatedAt.Compare(j1.CreatedAt)
	}, 0, 1000)
	if err != nil {
		return nil, err
	}
	for _, v := range vals {
		if jb, ok := v.(*job.Job); ok {
			jobs = append(jobs, jb)
		}
	}
	return jobs, nil
}
