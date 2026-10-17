package model

import "time"

// Backup describes one backup across local storage and Drive.
type Backup struct {
	ID              string     `json:"id" binding:"required"`
	LocalAvailable  bool       `json:"local_available" binding:"required"`
	DriveAvailable  bool       `json:"drive_available" binding:"required"`
	LocalSizeBytes  int64      `json:"local_size_bytes,omitempty"`
	DriveSizeBytes  int64      `json:"drive_size_bytes,omitempty"`
	LocalModifiedAt *time.Time `json:"local_modified_at,omitempty"`
	DriveModifiedAt *time.Time `json:"drive_modified_at,omitempty"`
}
