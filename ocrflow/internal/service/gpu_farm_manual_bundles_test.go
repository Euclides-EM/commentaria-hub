package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGPUFarmManualBundlesCleanupExpired(t *testing.T) {
	now := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	b := NewGPUFarmManualBundles("https://example.test/api", 24*time.Hour)
	b.root = t.TempDir()
	b.now = func() time.Time { return now }

	oldPath := writeManualBundleTestFile(t, b.root, "job_old", "run_261014-120000-1.zip", now.Add(-25*time.Hour))
	freshPath := writeManualBundleTestFile(t, b.root, "job_new", "run_261016-110000-1.zip", now.Add(-time.Hour))

	if err := b.CleanupExpired(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("expired bundle still exists: %v", err)
	}
	if _, err := os.Stat(freshPath); err != nil {
		t.Fatalf("fresh bundle missing: %v", err)
	}
}

func TestGPUFarmManualBundlesPathExpiresBundle(t *testing.T) {
	now := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	b := NewGPUFarmManualBundles("https://example.test/api", time.Hour)
	b.root = t.TempDir()
	b.now = func() time.Time { return now }

	path := writeManualBundleTestFile(t, b.root, "job_abc", "run_261016-100000-1.zip", now.Add(-2*time.Hour))
	if _, err := b.Path("job_abc", "run_261016-100000-1"); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("Path error = %v, want expired", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expired bundle still exists: %v", err)
	}
}

func writeManualBundleTestFile(t *testing.T, root, jobID, name string, modTime time.Time) string {
	t.Helper()
	path := filepath.Join(root, jobID, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("bundle"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
	return path
}
