package service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/gpufarm"
)

const gpuFarmManualBackend = "manual"

const defaultGPUFarmManualBundleTTL = 24 * time.Hour

var manualRunIDPattern = regexp.MustCompile(`^run_[0-9-]+$`)

var manualJobIDPattern = regexp.MustCompile(`^job_[a-z0-9]+$`)

type gpuFarmDispatch struct {
	JobID       string
	OnSubmitted func(string)
}

func (d *gpuFarmDispatch) jobID() string {
	if d == nil {
		return ""
	}
	return d.JobID
}

func (d *gpuFarmDispatch) submitted(message string) {
	if d != nil && d.OnSubmitted != nil {
		d.OnSubmitted(message)
	}
}

type GPUFarmManualBundles struct {
	apiURL string
	root   string
	ttl    time.Duration
	now    func() time.Time
}

func NewGPUFarmManualBundles(apiURL string, configuredTTL ...time.Duration) *GPUFarmManualBundles {
	ttl := defaultGPUFarmManualBundleTTL
	if len(configuredTTL) > 0 && configuredTTL[0] > 0 {
		ttl = configuredTTL[0]
	}
	return &GPUFarmManualBundles{
		apiURL: strings.TrimRight(apiURL, "/"),
		root:   filepath.Join(futils.TmpDir, "gpu-farm-manual-bundles"),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (b *GPUFarmManualBundles) Path(jobID string, runID string) (string, error) {
	if !manualJobIDPattern.MatchString(jobID) {
		return "", fmt.Errorf("invalid job ID %q", jobID)
	}
	if !manualRunIDPattern.MatchString(runID) {
		return "", fmt.Errorf("invalid GPU farm run ID %q", runID)
	}
	bundlePath := filepath.Join(b.root, jobID, runID+".zip")
	info, err := os.Stat(bundlePath)
	if err != nil {
		return "", err
	}
	if b.isExpired(info) {
		if err := os.Remove(bundlePath); err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("remove expired manual GPU farm bundle: %w", err)
		}
		_ = removeEmptyParent(filepath.Dir(bundlePath), b.root)
		return "", fmt.Errorf("manual GPU farm bundle expired after %s", b.ttl)
	}
	return bundlePath, nil
}

func (b *GPUFarmManualBundles) Publish(localJobDir string, jobID string, runID string, stageDir string) (string, error) {
	if b.apiURL == "" {
		return "", fmt.Errorf("API_URL is required for manual GPU farm runs")
	}
	if jobID == "" {
		return "", fmt.Errorf("manual GPU farm runs must be dispatched from an async job")
	}
	jobFiles := gpufarm.NewPythonEnvRequest(localJobDir)
	for _, filename := range jobFiles.Files {
		if err := futils.CopyFile(filepath.Join(jobFiles.LocalDir, filename), filepath.Join(stageDir, filename)); err != nil {
			return "", fmt.Errorf("copy job file %s to manual GPU farm bundle: %w", filename, err)
		}
	}
	if err := b.CleanupExpired(); err != nil {
		log.Printf("warning: manual GPU farm bundle cleanup before publish failed: %v", err)
	}
	if !manualJobIDPattern.MatchString(jobID) {
		return "", fmt.Errorf("invalid job ID %q", jobID)
	}
	if !manualRunIDPattern.MatchString(runID) {
		return "", fmt.Errorf("invalid GPU farm run ID %q", runID)
	}
	bundlePath := filepath.Join(b.root, jobID, runID+".zip")
	if err := os.MkdirAll(filepath.Dir(bundlePath), 0o700); err != nil {
		return "", err
	}
	if err := futils.Zip(stageDir, bundlePath); err != nil {
		return "", fmt.Errorf("write manual GPU farm bundle: %w", err)
	}
	return fmt.Sprintf("manual GPU farm run: follow docs/GPU_FARM.md#manual-gpu-farm-runs with\nJOB=%s\nRUN_ID=%s\nBUNDLE_URL=%s\n",
		envexec.ShellQuote(jobFiles.JobName),
		envexec.ShellQuote(runID),
		envexec.ShellQuote(b.apiURL+"/jobs/"+jobID+"/gpu_farm_runs/"+runID+"/bundle")), nil
}

func (b *GPUFarmManualBundles) isExpired(info fs.FileInfo) bool {
	return !b.now().Before(info.ModTime().Add(b.ttl))
}

// CleanupExpired removes bundles after their download TTL and prunes empty job
// directories. The dedicated root is intentionally outside the generic
// ocrflow-* startup cleanup so bundles remain available across API restarts.
func (b *GPUFarmManualBundles) CleanupExpired() error {
	if err := os.MkdirAll(b.root, 0o700); err != nil {
		return fmt.Errorf("create manual GPU farm bundle root: %w", err)
	}
	var cleanupErr error
	err := filepath.WalkDir(b.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".zip") {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !b.isExpired(info) {
			return nil
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("remove expired bundle %s: %w", path, err))
			return nil
		}
		log.Printf("removed expired manual GPU farm bundle: path=%s ttl=%s size=%d", path, b.ttl, info.Size())
		_ = removeEmptyParent(filepath.Dir(path), b.root)
		return nil
	})
	return errors.Join(err, cleanupErr)
}

func (b *GPUFarmManualBundles) StartCleanup(ctx context.Context) {
	interval := b.ttl / 4
	if interval > time.Hour {
		interval = time.Hour
	}
	if interval < time.Minute {
		interval = time.Minute
	}
	if err := b.CleanupExpired(); err != nil {
		log.Printf("warning: initial manual GPU farm bundle cleanup failed: %v", err)
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := b.CleanupExpired(); err != nil {
					log.Printf("warning: periodic manual GPU farm bundle cleanup failed: %v", err)
				}
			}
		}
	}()
}

func removeEmptyParent(dir string, stop string) error {
	if filepath.Clean(dir) == filepath.Clean(stop) {
		return nil
	}
	if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func manifestPathUnder(dirVar string, relPath string) string {
	if relPath == "" {
		return "''"
	}
	return fmt.Sprintf("\"$%s\"/%s", dirVar, envexec.ShellQuote(relPath))
}
