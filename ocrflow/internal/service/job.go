package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/job"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

type Job struct {
	jobsStore         *store.JobStore
	annotationsUpload *AnnotationsUploader
	facsimiles        *Facsimile
}

func (j *Job) ListJobs() ([]*job.Job, error) {
	return j.jobsStore.ListAll()
}

func (j *Job) ListIntegrationJobs() ([]*job.Job, error) {
	return j.ListJobs()
}

func (j *Job) CreateJob(ij *job.Job) (*job.Job, error) {
	res, err := j.CreateJobs(&job.Jobs{Jobs: []*job.Job{ij}})
	if err != nil {
		return nil, err
	}
	if len(res.Jobs) != 1 {
		return nil, fmt.Errorf("expected 1 job, got %d", len(res.Jobs))
	}
	return res.Jobs[0], nil
}

func (j *Job) CreateJobs(ij *job.Jobs) (*job.Jobs, error) {
	for _, jb := range ij.Jobs {
		j.createJob(jb)
		if jb.Task == job.FacsimileDriveImport {
			j.run(jb, "facsimile drive import", func() (any, error) {
				return j.facsimiles.ImportFromDriveInbox()
			})
		}
		if jb.Task == job.Export && jb.Target != nil && jb.Annotation != nil {
			switch jb.Target.Platform {
			case job.PlatformRoboflow:
				j.run(jb, "roboflow export", func() (any, error) {
					return nil, j.exportToRoboflow(jb)
				})
			case job.PlatformEScripturium:
				j.run(jb, "escriptorium export", func() (any, error) {
					return nil, j.exportToEScriptorium(jb)
				})
			case job.PlatformCommentaria:
				j.run(jb, "commentaria export", func() (any, error) {
					return nil, j.exportToCommentaria(jb)
				})
			}
		}
	}
	return ij, nil
}

func (j *Job) createJob(jb *job.Job) {
	jb.ID = idgen.GenerateID("job")
	jb.CreatedAt = time.Now()
	jb.UpdatedAt = jb.CreatedAt
	jb.Status = job.StatusPending
	j.jobsStore.Create(jb)
}

func (j *Job) exportToRoboflow(job *job.Job) error {
	rbu := &annotation.UploadRoboflow{
		APIKey:           job.Target.APIKey,
		WorkspaceID:      job.Target.WorkspaceID,
		ProjectID:        job.Target.ProjectID,
		IsNotGroundTruth: job.Target.IsNotGroundTruth,
	}
	_, err := j.annotationsUpload.UploadToRoboflow(job.Annotation.DatasetID, job.Annotation.ID, rbu)
	return err
}

func (j *Job) exportToEScriptorium(job *job.Job) error {
	ebu := &annotation.UploadEscriptorium{
		BasePath: job.Target.BasePath,
		Username: job.Target.Username,
		Password: job.Target.Password,
		Document: job.Target.Document,
	}
	_, err := j.annotationsUpload.UploadToEscriptorium(job.Annotation.DatasetID, job.Annotation.ID, ebu)
	return err
}

func (j *Job) exportToCommentaria(job *job.Job) error {
	cbu := &annotation.UploadCommentaria{
		BasePath:  job.Target.BasePath,
		APIKey:    job.Target.APIKey,
		DatasetID: job.Target.DatasetID,
	}
	_, err := j.annotationsUpload.UploadToCommentaria(job.Annotation.DatasetID, job.Annotation.ID, cbu)
	return err
}

func (j *Job) run(jb *job.Job, actionName string, action func() (any, error)) {
	go func() {
		now := time.Now()
		jb.Status = job.StatusRunning
		jb.UpdatedAt = now
		j.jobsStore.Update(jb)

		result, err := action()

		now = time.Now()
		jb.UpdatedAt = now
		jb.FinishedAt = &now
		if err != nil {
			jb.Status = job.StatusFailed
			jb.Details = err.Error()
			log.Printf("job %s %s failed: %v", jb.ID, actionName, err)
		} else {
			jb.Status = job.StatusCompleted
			if result != nil {
				if details, err := json.Marshal(result); err == nil {
					jb.Details = string(details)
				}
			}
		}
		j.jobsStore.Update(jb)
	}()
}

func (j *Job) GetIntegrationJob(id string) (*job.Job, error) {
	return j.jobsStore.Get(id)
}

func NewJob(jobsStore *store.JobStore, annotationsUpload *AnnotationsUploader, facsimiles *Facsimile) *Job {
	return &Job{
		jobsStore:         jobsStore,
		annotationsUpload: annotationsUpload,
		facsimiles:        facsimiles,
	}
}
