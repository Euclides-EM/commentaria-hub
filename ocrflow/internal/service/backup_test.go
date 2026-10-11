package service

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func newTestBackup(t *testing.T) (*Backup, string) {
	t.Helper()

	tmp := t.TempDir()
	baseDataDir := filepath.Join(tmp, "data")
	modelsDir := filepath.Join(tmp, "models")
	itemsDir := filepath.Join(tmp, "items")
	backupDir := filepath.Join(tmp, "backups")
	for _, dir := range []string{baseDataDir, modelsDir, itemsDir, backupDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(baseDataDir, "page.txt"), []byte("page"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modelsDir, "model.bin"), []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(itemsDir, "items.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmp, "ocrflow.db")
	if err := os.WriteFile(dbPath, []byte("database"), 0o644); err != nil {
		t.Fatal(err)
	}

	return NewBackup(baseDataDir, modelsDir, itemsDir, dbPath, backupDir, "", "", "", 5, nil), backupDir
}

func TestCreateBackupPromotesPartialOnlyAfterArchiveIsComplete(t *testing.T) {
	backup, backupDir := newTestBackup(t)
	backup.SetCheckpointFunc(func() error {
		partials, err := filepath.Glob(filepath.Join(backupDir, "*.zip.partial"))
		if err != nil {
			return err
		}
		if len(partials) != 1 {
			return errors.New("expected exactly one partial backup during creation")
		}
		listed, err := backup.ListBackups()
		if err != nil {
			return err
		}
		if len(listed) != 0 {
			return errors.New("partial backup was listed as complete")
		}
		return nil
	})

	id, err := backup.CreateBackup(false, nil)
	if err != nil {
		t.Fatal(err)
	}
	finalPath := filepath.Join(backupDir, id+".zip")
	if _, err := os.Stat(finalPath); err != nil {
		t.Fatalf("final backup does not exist: %v", err)
	}
	partials, err := filepath.Glob(filepath.Join(backupDir, "*.partial"))
	if err != nil {
		t.Fatal(err)
	}
	if len(partials) != 0 {
		t.Fatalf("partial backups remain after success: %v", partials)
	}
	zr, err := zip.OpenReader(finalPath)
	if err != nil {
		t.Fatalf("final backup is not a valid zip: %v", err)
	}
	defer zr.Close()
}

func TestCreateBackupRemovesPartialAfterError(t *testing.T) {
	backup, backupDir := newTestBackup(t)
	wantErr := errors.New("checkpoint failed")
	backup.SetCheckpointFunc(func() error { return wantErr })

	if _, err := backup.CreateBackup(false, nil); !errors.Is(err, wantErr) {
		t.Fatalf("CreateBackup error = %v, want %v", err, wantErr)
	}
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("backup files remain after failure: %v", entries)
	}
}

func TestAddDirToZipIncludesSymlinkedDirectoryContents(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src")
	targetDir := filepath.Join(tmp, "target")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "image.jpg"), []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetDir, filepath.Join(srcDir, "imgs")); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(tmp, "backup.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	if err := addDirToZip(zw, srcDir, "base_data"); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	for _, file := range zr.File {
		if file.Name == "base_data/imgs/image.jpg" {
			return
		}
	}
	t.Fatalf("zip does not contain symlinked directory file base_data/imgs/image.jpg")
}

func TestParseDriveBackupEntriesFiltersAndKeepsEntryType(t *testing.T) {
	got := parseDriveBackupEntries(strings.Join([]string{
		"euclides_backup_20260508T090000Z.zip",
		"euclides_backup_20260507T090000Z.zip/",
		"notes.txt",
		"nested/euclides_backup_20260506T090000Z.zip",
		"EUCLIDES_BACKUP_20260505T090000Z.ZIP",
		"",
	}, "\n"))

	want := []driveBackupEntry{
		{name: "euclides_backup_20260508T090000Z.zip"},
		{name: "euclides_backup_20260507T090000Z.zip", isDir: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %#v, want %#v", got, want)
	}
}
