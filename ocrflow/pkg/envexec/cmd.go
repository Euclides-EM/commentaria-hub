package envexec

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
)

func Cmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start yaltai: %w", err)
	}

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Printf("[%s stdout] %s", name, scanner.Text())
		}
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[%s stderr] %s", name, scanner.Text())
		}
	}()

	// Wait for exit
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s command failed: %w", name, err)
	}

	log.Printf("%s command succeeded. Args: %v", name, args)

	return nil
}
