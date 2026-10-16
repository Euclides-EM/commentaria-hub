package service

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
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

func TestParseDriveBackupMetadataFiltersAndKeepsMetadata(t *testing.T) {
	modifiedAt := time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC)
	data, err := json.Marshal([]rcloneJSONEntry{
		{Name: "euclides_backup_20260508T090000Z.zip", Size: 1234, ModTime: modifiedAt},
		{Name: "notes.txt", Size: 9, ModTime: modifiedAt},
		{Path: "nested/euclides_backup_20260507T090000Z.zip", Size: 10, ModTime: modifiedAt},
		{Name: "euclides_backup_20260506T090000Z.zip", IsDir: true, ModTime: modifiedAt},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := parseDriveBackupMetadata(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("entries = %#v, want one entry", got)
	}
	if got[0].name != "euclides_backup_20260508T090000Z.zip" || got[0].size != 1234 {
		t.Fatalf("entry = %#v, want backup name and size", got[0])
	}
	if got[0].modifiedAt == nil || !got[0].modifiedAt.Equal(modifiedAt) {
		t.Fatalf("modifiedAt = %v, want %v", got[0].modifiedAt, modifiedAt)
	}
}

func TestListBackupsIncludesLocalMetadata(t *testing.T) {
	backupDir := t.TempDir()
	name := "euclides_backup_20260508T090000Z.zip"
	if err := os.WriteFile(filepath.Join(backupDir, name), []byte("backup"), 0o644); err != nil {
		t.Fatal(err)
	}
	backup := NewBackup("", "", "", "", backupDir, "", "", "", 1, 5, nil)

	got, err := backup.ListBackups()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != strings.TrimSuffix(name, ".zip") {
		t.Fatalf("backups = %#v, want local backup", got)
	}
	if !got[0].LocalAvailable || got[0].DriveAvailable || got[0].LocalSizeBytes != 6 || got[0].LocalModifiedAt == nil {
		t.Fatalf("backup metadata = %#v, want local-only metadata", got[0])
	}
}

func TestZipUncompressedSize(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "backup.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, contents := range map[string]string{"one.txt": "1234", "two.txt": "56789"} {
		entry, createErr := writer.Create(name)
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, writeErr := entry.Write([]byte(contents)); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := zipUncompressedSize(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != 9 {
		t.Fatalf("uncompressed size = %d, want 9", got)
	}
}

func TestEnsureBackupDiskSpaceRejectsOversizedOperation(t *testing.T) {
	err := ensureBackupDiskSpace(t.TempDir(), math.MaxInt64, "test backup")
	if err == nil || !strings.Contains(err.Error(), "not enough disk space") {
		t.Fatalf("error = %v, want not enough disk space", err)
	}
}

func TestSetupRestoreBackupStagesDriveOnlyBackupAndCleansItUp(t *testing.T) {
	testRoot := t.TempDir()
	archivePath := filepath.Join(testRoot, "source.zip")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	zipWriter := zip.NewWriter(archive)
	entry, err := zipWriter.Create("base_data/restored.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("restored from Drive")); err != nil {
		t.Fatal(err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	archiveInfo, err := os.Stat(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	backupID := "euclides_backup_20260508T090000Z"
	listing, err := json.Marshal([]rcloneJSONEntry{{
		Name:    backupID + ".zip",
		Size:    archiveInfo.Size(),
		ModTime: time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC),
	}})
	if err != nil {
		t.Fatal(err)
	}
	fakeBinDir := filepath.Join(testRoot, "bin")
	if err := os.MkdirAll(fakeBinDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fakeRclone := filepath.Join(fakeBinDir, "rclone")
	script := "#!/bin/sh\n" +
		"case \" $* \" in\n" +
		"  *\" lsjson \"*) printf '%s' '" + string(listing) + "' ;;\n" +
		"  *) for arg in \"$@\"; do destination=\"$arg\"; done; cp \"$FAKE_BACKUP_ARCHIVE\" \"$destination\" ;;\n" +
		"esac\n"
	if err := os.WriteFile(fakeRclone, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_BACKUP_ARCHIVE", archivePath)

	oldTmpDir := futils.TmpDir
	stagingRoot := filepath.Join(testRoot, "staging")
	futils.SetTmpDir(stagingRoot)
	defer futils.SetTmpDir(oldTmpDir)

	restoreDir := filepath.Join(testRoot, "restore")
	backup := NewBackup("", "", "", "", filepath.Join(testRoot, "local"), restoreDir, "fake", "", 0, 5, nil)
	if err := backup.SetupRestoreBackup(backupID); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(filepath.Join(restoreDir, "base_data", "restored.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != "restored from Drive" {
		t.Fatalf("restored contents = %q", restored)
	}
	stagedEntries, err := os.ReadDir(stagingRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(stagedEntries) != 0 {
		t.Fatalf("temporary Drive staging files remain: %#v", stagedEntries)
	}
}

func TestNewBackupUsesSeparateRetentionDefaults(t *testing.T) {
	backup := NewBackup("", "", "", "", "", "", "", "", 0, 0, nil)

	if backup.localMaxBackupsToStore != 0 {
		t.Fatalf("localMaxBackupsToStore = %d, want 0", backup.localMaxBackupsToStore)
	}
	if backup.driveMaxBackupsToStore != 5 {
		t.Fatalf("driveMaxBackupsToStore = %d, want 5", backup.driveMaxBackupsToStore)
	}
}

func TestCreateBackupRequiresDriveSyncWhenLocalRetentionDisabled(t *testing.T) {
	backup := NewBackup("", "", "", "", t.TempDir(), "", "", "", 0, 5, nil)
	if _, err := backup.CreateBackup(false, nil); err == nil || !strings.Contains(err.Error(), "sync_to_drive must be true") {
		t.Fatalf("CreateBackup error = %v, want sync_to_drive requirement", err)
	}
}

func TestCreateBackupFromZipRequiresStorageDestination(t *testing.T) {
	backup := NewBackup("", "", "", "", t.TempDir(), "", "", "", 0, 5, nil)
	_, err := backup.CreateBackupFromZip(nopMultipartFile{Reader: bytes.NewReader(nil)}, func(string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "backup Drive is not configured") {
		t.Fatalf("CreateBackupFromZip error = %v, want missing Drive configuration", err)
	}
}

func TestCreateBackupFromZipUsesTemporaryValidationDirectory(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	entry, err := writer.Create("base_data/example.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("example")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	backupDir := t.TempDir()
	backup := NewBackup("", "", "", "", backupDir, "", "", "", 1, 5, nil)
	var validationDir string
	id, err := backup.CreateBackupFromZip(nopMultipartFile{Reader: bytes.NewReader(archive.Bytes())}, func(path string) error {
		validationDir = path
		info, statErr := os.Stat(path)
		if statErr != nil {
			return statErr
		}
		if !info.IsDir() {
			return errors.New("validation path is not a directory")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(backupDir, id+".zip")); err != nil {
		t.Fatalf("stored backup: %v", err)
	}
	if _, err := os.Stat(validationDir); !os.IsNotExist(err) {
		t.Fatalf("validation directory was not cleaned up: %v", err)
	}
}

type nopMultipartFile struct {
	*bytes.Reader
}

func (nopMultipartFile) Close() error { return nil }

func TestEnsureMaxBackupsAllowsZeroLocalRetention(t *testing.T) {
	backupDir := t.TempDir()
	name := "euclides_backup_20260508T090000Z.zip"
	if err := os.WriteFile(filepath.Join(backupDir, name), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	backup := NewBackup("", "", "", "", backupDir, "", "", "", 0, 5, nil)
	if err := backup.ensureMaxBackups(); err != nil {
		t.Fatal(err)
	}
	got, err := backup.ListLocalBackups()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("backups = %#v, want none", got)
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

	got, err := backup.ListLocalBackups()
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
