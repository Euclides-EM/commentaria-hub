package gpufarm

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
)

var slurmJobIDPattern = regexp.MustCompile(`Submitted batch job\s+([^\s]+)`)

const completedRunsToKeep = 3

//go:embed templates/cleanup_completed_runs.sh
var cleanupCompletedRunsScript string

//go:embed templates/python_venv_setup.sh
var pythonVEnvSetupScript string

var cleanupCompletedRunsTemplate = template.
	Must(template.New("cleanup-completed-gpu-runs").
		Funcs(template.FuncMap{
			"shellquote": envexec.ShellQuote,
		}).Parse(cleanupCompletedRunsScript))

type cleanupCompletedRunsTemplateData struct {
	JobRoot             string
	CompletedRunsToKeep int
}

func renderCleanupCompletedRunsScript(jobRoot string, runsToKeep int) (string, error) {
	var script bytes.Buffer
	if err := cleanupCompletedRunsTemplate.Execute(&script, cleanupCompletedRunsTemplateData{
		JobRoot:             jobRoot,
		CompletedRunsToKeep: runsToKeep,
	}); err != nil {
		return "", err
	}
	return script.String(), nil
}

type SubmitterSlurm struct {
	host    string
	jobRoot string
}

func NewSubmitterSlurm(host string, jobRoot string) *SubmitterSlurm {
	return &SubmitterSlurm{host: host, jobRoot: jobRoot}
}

func (s *SubmitterSlurm) run(args ...string) (string, error) {
	remoteCommand := make([]string, 0, len(args))
	for _, arg := range args {
		remoteCommand = append(remoteCommand, envexec.ShellQuote(arg))
	}
	sshArgs := []string{"-o", "StrictHostKeyChecking=accept-new", s.host, strings.Join(remoteCommand, " ")}
	out, err := exec.Command("ssh", sshArgs...).CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if output != "" {
			return output, fmt.Errorf("ssh remote job host: %w: %s", err, output)
		}
		return output, fmt.Errorf("ssh remote job host: %w", err)
	}
	return output, nil
}

func (s *SubmitterSlurm) CopyTo(localPath string, remotePath string) error {
	if _, err := s.run("mkdir", "-p", path.Dir(remotePath)); err != nil {
		return fmt.Errorf("create directory %s remotely: %w", path.Dir(remotePath), err)
	}
	out, err := exec.Command("scp", "-o", "StrictHostKeyChecking=accept-new", localPath, fmt.Sprintf("%s:%s", s.host, remotePath)).CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if output != "" {
			return fmt.Errorf("scp to remote job host: %w: %s", err, output)
		}
		return fmt.Errorf("scp to remote job host: %w", err)
	}
	return nil
}

func (s *SubmitterSlurm) FileExists(remotePath string) (bool, error) {
	output, err := s.run("bash", "-lc", fmt.Sprintf("if [[ -f %s ]]; then echo exists; else echo missing; fi", envexec.ShellQuote(remotePath)))
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) == "exists", nil
}

// cleanupCompletedRuns bounds disk usage across every managed job kind. Slurm
// work directories are protected even when they are older than the retained
// completed runs.
func (s *SubmitterSlurm) cleanupCompletedRuns() error {
	script, err := renderCleanupCompletedRunsScript(s.jobRoot, completedRunsToKeep)
	if err != nil {
		return fmt.Errorf("render completed run cleanup script: %w", err)
	}

	output, err := s.run("bash", "-lc", script)
	if output != "" {
		log.Printf("GPU farm cleanup:\n%s", output)
	}
	if err != nil {
		return fmt.Errorf("clean up completed runs in %s: %w", s.jobRoot, err)
	}
	return nil
}

func (s *SubmitterSlurm) PreparePythonEnv(request PythonEnvRequest) (*RemoteEnv, error) {
	if strings.TrimSpace(request.JobName) == "" {
		return nil, fmt.Errorf("missing Python job directory")
	}
	if strings.TrimSpace(request.LocalDir) == "" {
		return nil, fmt.Errorf("missing local Python job directory")
	}
	if len(request.Files) == 0 {
		return nil, fmt.Errorf("missing Python job files")
	}

	runID := "run_" + fmt.Sprintf("%s-%d", time.Now().UTC().Format("060102-150405"), time.Now().UnixNano()%100000)
	remoteRoot := path.Join(s.jobRoot, request.JobName)
	if _, err := s.run("mkdir", "-p", s.jobRoot); err != nil {
		return nil, err
	}
	if err := s.cleanupCompletedRuns(); err != nil {
		return nil, err
	}
	if _, err := s.run("mkdir", "-p", remoteRoot); err != nil {
		return nil, err
	}
	remoteRunDir := path.Join(remoteRoot, runID)
	logDir := path.Join(remoteRunDir, "logs")
	if _, err := s.run("mkdir", "-p", remoteRunDir, logDir); err != nil {
		return nil, err
	}

	for _, filename := range request.Files {
		localPath := filepath.Join(request.LocalDir, filename)
		if err := s.CopyTo(localPath, path.Join(remoteRoot, filename)); err != nil {
			return nil, err
		}
	}

	envScript := fmt.Sprintf(pythonVEnvSetupScript, envexec.ShellQuote(remoteRoot))
	if _, err := s.run("bash", "-lc", envScript); err != nil {
		return nil, err
	}
	return &RemoteEnv{
		RemoteDir:    remoteRoot,
		RemoteRunDir: remoteRunDir,
		RunID:        runID,
		LogsDir:      logDir,
	}, nil
}

func (s *SubmitterSlurm) Submit(request *RemoteEnv) (*JobSubmission, error) {
	if strings.TrimSpace(request.RemoteDir) == "" {
		return nil, fmt.Errorf("missing remote run directory")
	}
	if strings.TrimSpace(path.Join(request.RemoteDir, "job.sbatch")) == "" {
		return nil, fmt.Errorf("missing Slurm entrypoint")
	}

	output, err := s.run("bash", "-lc", fmt.Sprintf("cd %s && sbatch %s", envexec.ShellQuote(request.RemoteRunDir), envexec.ShellQuote(path.Join(request.RemoteDir, "job.sbatch"))))
	if err != nil {
		return nil, err
	}

	submission := &JobSubmission{
		Backend:      "slurm",
		Host:         s.host,
		SubmitOutput: strings.TrimSpace(output),
	}
	if match := slurmJobIDPattern.FindStringSubmatch(output); len(match) == 2 {
		submission.SchedulerJobID = match[1]
	}
	return submission, nil
}
