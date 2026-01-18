package envexec

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Cmd(name string, args ...string) error {
	return CmdWithEnv(nil, name, args...)
}

func CmdWithEnv(env map[string]string, name string, args ...string) error {
	cmd := exec.Command(name, args...)

	if env != nil && len(env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", name, err)
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Printf("[%s stdout] %s", name, scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[%s stderr] %s", name, scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s command failed: %w", name, err)
	}

	log.Printf("%s command succeeded. Args: %v", name, args)

	return nil
}
