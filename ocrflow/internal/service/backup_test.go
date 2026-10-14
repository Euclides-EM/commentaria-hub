package service

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

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

func TestNewBackupUsesSeparateRetentionDefaults(t *testing.T) {
	backup := NewBackup("", "", "", "", "", "", "", "", 0, 0, nil)

	if backup.localMaxBackupsToStore != 2 {
		t.Fatalf("localMaxBackupsToStore = %d, want 2", backup.localMaxBackupsToStore)
	}
	if backup.driveMaxBackupsToStore != 5 {
		t.Fatalf("driveMaxBackupsToStore = %d, want 5", backup.driveMaxBackupsToStore)
	}
}

func TestEnsureMaxBackupsUsesLocalRetention(t *testing.T) {
	backupDir := t.TempDir()
	for _, name := range []string{
		"euclides_backup_20260508T090000Z.zip",
		"euclides_backup_20260507T090000Z.zip",
		"euclides_backup_20260506T090000Z.zip",
	} {
		if err := os.WriteFile(filepath.Join(backupDir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	backup := NewBackup("", "", "", "", backupDir, "", "", "", 2, 5, nil)
	if err := backup.ensureMaxBackups(); err != nil {
		t.Fatal(err)
	}

	got, err := backup.ListBackups()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"euclides_backup_20260508T090000Z",
		"euclides_backup_20260507T090000Z",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("backups = %#v, want %#v", got, want)
	}
}
