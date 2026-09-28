package gpufarm

import (
	"fmt"
	"path"
	"time"
)

type RemoteEnv struct {
	RemoteDir    string
	RemoteRunDir string
	RunID        string
	LogsDir      string
}

func NewRunID() string {
	return "run_" + fmt.Sprintf("%s-%d", time.Now().UTC().Format("060102-150405"), time.Now().UnixNano()%100000)
}

type PythonEnvRequest struct {
	JobName  string
	LocalDir string
	Files    []string
}

func NewPythonEnvRequest(localDir string) PythonEnvRequest {
	return PythonEnvRequest{
		JobName:  path.Base(localDir),
		LocalDir: localDir,
		Files:    []string{"script.py", "requirements.txt", "job.sbatch"},
	}
}

type JobSubmission struct {
	Backend        string
	Host           string
	SubmitOutput   string
	SchedulerJobID string
}

type Submitter interface {
	CopyTo(localPath string, remotePath string) error
	Discard(request *RemoteEnv) error
	FileExists(remotePath string) (bool, error)
	PreparePythonEnv(request PythonEnvRequest) (*RemoteEnv, error)
	Submit(request *RemoteEnv) (*JobSubmission, error)
}
