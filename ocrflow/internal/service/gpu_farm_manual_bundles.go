package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/gpufarm"
)

const gpuFarmManualBackend = "manual"

var manualRunIDPattern = regexp.MustCompile(`^run_[0-9-]+$`)

type GPUFarmManualBundles struct {
	apiURL string
}

func NewGPUFarmManualBundles(apiURL string) *GPUFarmManualBundles {
	return &GPUFarmManualBundles{apiURL: strings.TrimRight(apiURL, "/")}
}

func (b *GPUFarmManualBundles) Path(runID string) (string, error) {
	if !manualRunIDPattern.MatchString(runID) {
		return "", fmt.Errorf("invalid GPU farm run ID %q", runID)
	}
	return filepath.Join(futils.TmpDir, "ocrflow-gpu-farm-manual", runID+".zip"), nil
}

func (b *GPUFarmManualBundles) Publish(localJobDir string, runID string, stageDir string) (string, error) {
	if b.apiURL == "" {
		return "", fmt.Errorf("API_URL is required for manual GPU farm runs")
	}
	jobFiles := gpufarm.NewPythonEnvRequest(localJobDir)
	for _, filename := range jobFiles.Files {
		if err := futils.CopyFile(filepath.Join(jobFiles.LocalDir, filename), filepath.Join(stageDir, filename)); err != nil {
			return "", fmt.Errorf("copy job file %s to manual GPU farm bundle: %w", filename, err)
		}
	}
	bundlePath, err := b.Path(runID)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(bundlePath), 0o700); err != nil {
		return "", fmt.Errorf("create manual GPU farm bundle directory: %w", err)
	}
	if err := futils.Zip(stageDir, bundlePath); err != nil {
		return "", fmt.Errorf("write manual GPU farm bundle: %w", err)
	}
	return fmt.Sprintf("manual GPU farm run: follow docs/GPU_FARM.md#manual-gpu-farm-runs with JOB=%s RUN_ID=%s BUNDLE_URL=%s",
		envexec.ShellQuote(jobFiles.JobName),
		envexec.ShellQuote(runID),
		envexec.ShellQuote(b.apiURL+"/gpu_farm/manual_bundles/"+runID)), nil
}

func manifestPathUnder(dirVar string, relPath string) string {
	if relPath == "" {
		return "''"
	}
	return fmt.Sprintf("\"$%s\"/%s", dirVar, envexec.ShellQuote(relPath))
}
