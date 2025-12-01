package krakenwrapper

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
)

// yolo train resume model=path/to/last.pt
func TrainYOLOModel(originModelLocalPath, datsetYmlPath, outputPath string) (<-chan error, error) {
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output dir %q: %w", outputPath, err)
	}

	args := []string{
		"detect",
		"train",
		fmt.Sprintf("data=%s", datsetYmlPath),
		fmt.Sprintf("model=%s", originModelLocalPath),
		//"epochs=100",
		//"imgsz=640",
		"save=True",
		fmt.Sprintf("project=%s", path.Dir(outputPath)),
		fmt.Sprintf("name=%s", path.Base(outputPath)),
	}

	name := "yolo"
	cmd := exec.Command(name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	errCh := make(chan error, 1)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start yolo: %w", err)
	}

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Printf("[yolo stdout] %s", scanner.Text())
		}
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[yolo stderr] %s", scanner.Text())
		}
	}()

	// Wait for completion and send final err
	go func() {
		defer close(errCh)
		if err := cmd.Wait(); err != nil {
			errCh <- fmt.Errorf("yolo training failed: %w", err)
			return
		}
		errCh <- nil
	}()

	return errCh, nil
}
