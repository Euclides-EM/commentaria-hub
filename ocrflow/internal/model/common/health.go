package common

type HealthStatus struct {
	DBReady   bool   `json:"db_ready"`
	CommitSHA string `json:"commit_sha,omitempty"`
}
