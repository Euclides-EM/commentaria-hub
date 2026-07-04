package service

import (
	"fmt"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/job"
)

type ModelTraining struct {
	jobs *Job
}

func NewModelTraining(jobs *Job) *ModelTraining {
	return &ModelTraining{jobs: jobs}
}

func (m *ModelTraining) Submit(training *model.ModelTraining) (*model.ModelTraining, error) {
	if training == nil {
		return nil, fmt.Errorf("missing model training request")
	}
	if training.Model == nil {
		return nil, fmt.Errorf("model training request is missing model")
	}
	createdJob, err := m.jobs.CreateJob(&job.Job{
		Task:          job.ModelTrain,
		ModelTraining: training,
	})
	if err != nil {
		return nil, err
	}
	return &model.ModelTraining{
		Meta:   common.NewMeta(createdJob.ID).WithCreatedAt(createdJob.CreatedAt).WithUpdatedAt(createdJob.UpdatedAt),
		Model:  createdJob.ModelTraining.Model,
		Epochs: training.Epochs,
		Status: model.ModelTrainingStatusSubmitted,
		StatusDetails: map[string]string{
			"job_id": createdJob.ID,
		},
	}, nil
}
