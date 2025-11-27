package roboflow

import (
	_ "embed"
	"fmt"
	"github.com/samber/lo"
	"log"
	"os"
	"os/exec"
	"strings"
)

//go:embed upload_dataset.py
var pythonScriptUploadDataset string

type UploadDatasetParams struct {
	APIKey           string
	WorkspaceID      string
	DatasetPath      string
	ProjectID        string
	IsNotGroundTruth bool
}

func NewUploadDatasetParams() *UploadDatasetParams {
	return &UploadDatasetParams{}
}

func (p *UploadDatasetParams) SetAPIKey(apiKey string) *UploadDatasetParams {
	p.APIKey = apiKey
	return p
}

func (p *UploadDatasetParams) SetWorkspaceID(workspaceID string) *UploadDatasetParams {
	p.WorkspaceID = workspaceID
	return p
}

func (p *UploadDatasetParams) SetDatasetPath(datasetPath string) *UploadDatasetParams {
	p.DatasetPath = datasetPath
	return p
}

func (p *UploadDatasetParams) SetProjectID(projectID string) *UploadDatasetParams {
	p.ProjectID = projectID
	return p
}

func (p *UploadDatasetParams) SetIsNotGroundTruth(isNotGroundTruth bool) *UploadDatasetParams {
	p.IsNotGroundTruth = isNotGroundTruth
	return p
}

func UploadDataset(pythonExecutable string, p *UploadDatasetParams) error {
	script := pythonScriptUploadDataset
	script = strings.ReplaceAll(script, "API_KEY", p.APIKey)
	script = strings.ReplaceAll(script, "WORKSPACE_ID", p.WorkspaceID)
	script = strings.ReplaceAll(script, "DATASET_PATH", p.DatasetPath)
	script = strings.ReplaceAll(script, "PROJECT_ID", p.ProjectID)
	script = strings.ReplaceAll(script, "IS_NOT_GROUND_TRUTH", lo.Ternary(p.IsNotGroundTruth, "True", "False"))

	tmp, err := os.CreateTemp("", "script-*.py")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	tmp.WriteString(script)
	tmp.Close()

	cmd := exec.Command(pythonExecutable, tmp.Name())

	cmd.Stdin = strings.NewReader(script)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("python error: %v\n%s", err, out)
	}

	log.Printf("Python output:\n%s", out)
	return nil
}
