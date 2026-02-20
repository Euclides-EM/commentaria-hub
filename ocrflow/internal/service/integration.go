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
		if job.Task == integration.Export && job.Target != nil && job.Annotation != nil {
			switch job.Target.Platform {
			case integration.PlatformRoboflow:
				j.runExport(job, j.exportToRoboflow)
			case integration.PlatformEScripturium:
				j.runExport(job, j.exportToEScriptorium)
			case integration.PlatformCommentaria:
				j.runExport(job, j.exportToCommentaria)
			}
		}
	}
	return ij, nil
}

func (j *IntegrationJob) exportToRoboflow(job *integration.Job) error {
	rbu := &annotation.UploadRoboflow{
		APIKey:           job.Target.APIKey,
		WorkspaceID:      job.Target.WorkspaceID,
		ProjectID:        job.Target.ProjectID,
		IsNotGroundTruth: job.Target.IsNotGroundTruth,
	}
	_, err := j.annotationsUpload.UploadToRoboflow(job.Annotation.DatasetID, job.Annotation.ID, rbu)
	return err
}

func (j *IntegrationJob) exportToEScriptorium(job *integration.Job) error {
	ebu := &annotation.UploadEscriptorium{
		BasePath: job.Target.BasePath,
		Username: job.Target.Username,
		Password: job.Target.Password,
		Document: job.Target.Document,
	}
	_, err := j.annotationsUpload.UploadToEscriptorium(job.Annotation.DatasetID, job.Annotation.ID, ebu)
	return err
}

func (j *IntegrationJob) exportToCommentaria(job *integration.Job) error {
	cbu := &annotation.UploadCommentaria{
		BasePath:  job.Target.BasePath,
		APIKey:    job.Target.APIKey,
		DatasetID: job.Target.DatasetID,
	}
	_, err := j.annotationsUpload.UploadToCommentaria(job.Annotation.DatasetID, job.Annotation.ID, cbu)
	return err
}

func (j *IntegrationJob) runExport(job *integration.Job, export func(job *integration.Job) error) {
	go func() {
		now := time.Now()
		job.Status = integration.JobStatusRunning
		job.UpdatedAt = now
		j.jobsStore.Update(job)

		err := export(job)

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
