package common

type HealthStatus struct {
	DBReady         bool              `json:"db_ready"`
	CommitSHA       string            `json:"commit_sha,omitempty"`
	Filesystem      *FilesystemStatus `json:"filesystem,omitempty"`
	FilesystemError string            `json:"filesystem_error,omitempty"`
}

type FilesystemStatus struct {
	Path           string  `json:"path"`
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsedPercent    float64 `json:"used_percent"`
}
