package service

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type RcloneDrive struct {
	RemoteName        string
	DriveRootFolderID string
	Label             string
}

func NewRcloneDrive(remoteName, driveRootFolderID, label string) *RcloneDrive {
	return &RcloneDrive{
		RemoteName:        strings.TrimSpace(remoteName),
		DriveRootFolderID: strings.TrimSpace(driveRootFolderID),
		Label:             strings.TrimSpace(label),
	}
}

func (d *RcloneDrive) RemotePath(path string) string {
	return d.RemoteName + ":" + path
}

func (d *RcloneDrive) Run(args ...string) ([]byte, error) {
	if d.DriveRootFolderID != "" {
		args = append([]string{"--drive-root-folder-id=" + d.DriveRootFolderID}, args...)
	}
	label := d.Label
	if label == "" {
		label = "rclone"
	}
	log.Printf("running %s %s", label, strings.Join(args, " "))
	cmd := exec.Command("rclone", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("rclone: %w: %s", err, strings.TrimSpace(string(out)))
	}
	log.Printf("%s completed: %s", label, strings.TrimSpace(string(out)))
	return out, nil
}
