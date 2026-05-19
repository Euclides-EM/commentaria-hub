package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
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

func (d *RcloneDrive) RunStreaming(progress func(string), args ...string) ([]byte, error) {
	if d.DriveRootFolderID != "" {
		args = append([]string{"--drive-root-folder-id=" + d.DriveRootFolderID}, args...)
	}
	label := d.Label
	if label == "" {
		label = "rclone"
	}
	log.Printf("running %s %s", label, strings.Join(args, " "))
	cmd := exec.Command("rclone", args...)
	out, err := runWithStreamingOutput(cmd, progress)
	if err != nil {
		return nil, fmt.Errorf("rclone: %w: %s", err, strings.TrimSpace(string(out)))
	}
	log.Printf("%s completed: %s", label, strings.TrimSpace(string(out)))
	return out, nil
}

func runWithStreamingOutput(cmd *exec.Cmd, progress func(string)) ([]byte, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	var mu sync.Mutex
	var wg sync.WaitGroup
	read := func(r io.Reader) {
		defer wg.Done()
		readStreamingOutput(r, func(chunk []byte) {
			mu.Lock()
			out.Write(chunk)
			mu.Unlock()
			reportStreamingProgress(chunk, progress)
		})
	}
	wg.Add(2)
	go read(stdout)
	go read(stderr)
	wg.Wait()

	err = cmd.Wait()
	mu.Lock()
	defer mu.Unlock()
	return append([]byte(nil), out.Bytes()...), err
}

func readStreamingOutput(r io.Reader, handle func([]byte)) {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			handle(buf[:n])
		}
		if err != nil {
			return
		}
	}
}

func reportStreamingProgress(chunk []byte, progress func(string)) {
	for _, part := range strings.FieldsFunc(string(chunk), func(r rune) bool {
		return r == '\n' || r == '\r'
	}) {
		line := strings.TrimSpace(part)
		if line != "" {
			progress(line)
		}
	}
}
