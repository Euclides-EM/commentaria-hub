package roboflow

import (
	"path/filepath"
	"runtime"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
	"github.com/samber/lo"
)

const roboflowAPIURL = "https://api.roboflow.com"

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
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	scriptPath := filepath.Join(rootDir, "python-tools", "roboflow_upload_dataset.py")

	return envexec.PythonCmdWithEnv(uploadDatasetEnv(p), "python", scriptPath)
}

func uploadDatasetEnv(p *UploadDatasetParams) map[string]string {
	return map[string]string{
		// The Roboflow SDK uses the generic API_URL variable for its own API.
		// Override the Commentaria service's API_URL in this child process.
		"API_URL":                      roboflowAPIURL,
		"ROBOFLOW_API_KEY":             p.APIKey,
		"ROBOFLOW_WORKSPACE_ID":        p.WorkspaceID,
		"ROBOFLOW_DATASET_PATH":        p.DatasetPath,
		"ROBOFLOW_PROJECT_ID":          p.ProjectID,
		"ROBOFLOW_IS_NOT_GROUND_TRUTH": lo.Ternary(p.IsNotGroundTruth, "True", "False"),
	}
}
