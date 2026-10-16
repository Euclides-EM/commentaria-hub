package service

import "testing"

func TestFilesystemStatus(t *testing.T) {
	status, errMessage := filesystemStatus(t.TempDir())
	if errMessage != "" {
		t.Fatalf("filesystemStatus returned error: %s", errMessage)
	}
	if status == nil {
		t.Fatal("expected filesystem status")
	}
	if status.TotalBytes == 0 {
		t.Fatal("expected non-zero filesystem capacity")
	}
	if status.UsedBytes > status.TotalBytes {
		t.Fatalf("used bytes %d exceed total bytes %d", status.UsedBytes, status.TotalBytes)
	}
	if status.UsedPercent < 0 || status.UsedPercent > 100 {
		t.Fatalf("used percent out of range: %f", status.UsedPercent)
	}
}

func TestFilesystemStatusError(t *testing.T) {
	status, errMessage := filesystemStatus(t.TempDir() + "/missing")
	if status != nil {
		t.Fatal("expected no filesystem status for missing path")
	}
	if errMessage == "" {
		t.Fatal("expected filesystem error message")
	}
}
