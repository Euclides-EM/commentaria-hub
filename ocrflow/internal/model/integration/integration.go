package integration

import (
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type Platform string

const (
	PlatformRoboflow     Platform = "Roboflow"
	PlatformEScripturium Platform = "EScriptorium"
	PlatformCommentaria  Platform = "Commentaria"
)

var AllTypes = []Platform{PlatformRoboflow, PlatformEScripturium, PlatformCommentaria}

type Task string

const (
	Export Task = "Export"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type Job struct {
	common.Meta `json:",inline"`
	Task        Task                  `json:"task"`
	Target      *JobTarget            `json:"target"`
	Annotation  *annotation.Reference `json:"annotation,omitempty"`
	Status      JobStatus             `json:"status" readonly:"true"`
	FinishedAt  *time.Time            `json:"finished_at,omitempty" readonly:"true"`
	Details     string                `json:"details,omitempty"  readonly:"true"`
}

type Jobs struct {
	Jobs []*Job `json:"jobs"`
}

type JobTarget struct {
	Platform Platform `json:"platform"`

	// For EScriptorium
	Username string `json:"username"`
	Password string `json:"password"`
	BasePath string `json:"base_path"`
	Document string `json:"document"`

	// For Roboflow and Commentaria
	APIKey string `json:"api_key"`

	// For Roboflow
	WorkspaceID      string `json:"workspace_url" example:"mia-workplace"`
	ProjectID        string `json:"project_id" example:"dec06miamia-afl6i"`
	IsNotGroundTruth bool   `json:"is_not_ground_truth"`

	// For Commentaria
	DatasetID string `json:"dataset_id"`
}
