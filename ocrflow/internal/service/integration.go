package service

import (
	"log"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/integration"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

type IntegrationJob struct {
	jobsStore         *store.IntegrationJobStore
	annotationsUpload *AnnotationsUploader
}

func (j *IntegrationJob) ListIntegrationJobs() ([]*integration.Job, error) {
	return j.jobsStore.ListAll()
}

func (j *IntegrationJob) CreateIntegrationJobs(ij *integration.Jobs) (*integration.Jobs, error) {
	for _, job := range ij.Jobs {
		job.ID = idgen.GenerateID("job")
		job.CreatedAt = time.Now()
		job.UpdatedAt = job.CreatedAt
		job.Status = integration.JobStatusPending
		j.jobsStore.Create(job)
		if job.Task == integration.Export && job.Target != nil && job.Target.Platform == integration.PlatformRoboflow && job.Annotation != nil {
			j.runRoboflowExport(job)
		}
	}
	return ij, nil
}

func (j *IntegrationJob) runRoboflowExport(job *integration.Job) {
	go func() {
		now := time.Now()
		job.Status = integration.JobStatusRunning
		job.UpdatedAt = now
		j.jobsStore.Update(job)

		rbu := &annotation.UploadRoboflow{
			APIKey:           job.Target.APIKey,
			WorkspaceID:      job.Target.WorkspaceID,
			ProjectID:        job.Target.ProjectID,
			IsNotGroundTruth: job.Target.IsNotGroundTruth,
		}
		_, err := j.annotationsUpload.UploadToRoboflow(job.Annotation.DatasetID, job.Annotation.ID, rbu)
		now = time.Now()
		job.UpdatedAt = now
		job.FinishedAt = &now
		if err != nil {
			job.Status = integration.JobStatusFailed
			job.Details = err.Error()
			log.Printf("integration job %s roboflow export failed: %v", job.ID, err)
		} else {
			job.Status = integration.JobStatusCompleted
		}
		j.jobsStore.Update(job)
	}()
}

func (j *IntegrationJob) GetIntegrationJob(id string) (*integration.Job, error) {
	return j.jobsStore.Get(id)
}

func NewIntegrationJob(jobsStore *store.IntegrationJobStore, annotationsUpload *AnnotationsUploader) *IntegrationJob {
	return &IntegrationJob{
		jobsStore:         jobsStore,
		annotationsUpload: annotationsUpload,
	}
}
