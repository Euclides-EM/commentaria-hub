package roboflow

import "testing"

func TestUploadDatasetEnvOverridesAPIURL(t *testing.T) {
	p := NewUploadDatasetParams().
		SetAPIKey("test-key").
		SetWorkspaceID("test-workspace").
		SetDatasetPath("/tmp/test-dataset").
		SetProjectID("test-project").
		SetIsNotGroundTruth(true)

	env := uploadDatasetEnv(p)

	if got := env["API_URL"]; got != "https://api.roboflow.com" {
		t.Fatalf("API_URL = %q, want Roboflow API URL", got)
	}
	if got := env["ROBOFLOW_API_KEY"]; got != "test-key" {
		t.Errorf("ROBOFLOW_API_KEY = %q, want test-key", got)
	}
	if got := env["ROBOFLOW_WORKSPACE_ID"]; got != "test-workspace" {
		t.Errorf("ROBOFLOW_WORKSPACE_ID = %q, want test-workspace", got)
	}
	if got := env["ROBOFLOW_DATASET_PATH"]; got != "/tmp/test-dataset" {
		t.Errorf("ROBOFLOW_DATASET_PATH = %q, want /tmp/test-dataset", got)
	}
	if got := env["ROBOFLOW_PROJECT_ID"]; got != "test-project" {
		t.Errorf("ROBOFLOW_PROJECT_ID = %q, want test-project", got)
	}
	if got := env["ROBOFLOW_IS_NOT_GROUND_TRUTH"]; got != "True" {
		t.Errorf("ROBOFLOW_IS_NOT_GROUND_TRUTH = %q, want True", got)
	}
}
