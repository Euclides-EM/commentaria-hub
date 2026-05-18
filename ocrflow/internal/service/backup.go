package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/samber/lo"
)

const (
	defaultMaxBackupsToStore = 5
	zipProgressLogEveryFiles = 100
	backupZipPrefix          = "euclides_backup_"
)

type dirZipStats struct {
	files int
	dirs  int
	size  int64
}

type driveBackupEntry struct {
	name  string
	isDir bool
}

type Backup struct {
	baseDataDir      string
	modelsDir        string
	itemsMetadataDir string
	dbPath           string

	backupdir             string
	restoreFromBackupPath string

	rclone            *RcloneDrive
	maxBackupsToStore int

	shutdownFunc func() error
	// checkpointDB, if set, is called before copying the DB file into the backup.
	// It should run PRAGMA wal_checkpoint(FULL) so the main .db file contains all data
	// (otherwise in WAL mode only the main file is copied and recent writes are missing).
	checkpointDB func() error
}

func NewBackup(baseData, models, items, db, backupDir, restoreDir, rcloneRemoteName, rcloneGDriveFolderID string, maxBackupsToStore int, shutdownFunc func() error) *Backup {
	if maxBackupsToStore <= 0 {
		maxBackupsToStore = defaultMaxBackupsToStore
	}
	log.Printf("Creating new backup service with max backups to store: %d", maxBackupsToStore)
	return &Backup{
		baseDataDir:           baseData,
		modelsDir:             models,
		itemsMetadataDir:      items,
		dbPath:                db,
		backupdir:             backupDir,
		restoreFromBackupPath: restoreDir,
		rclone:                NewRcloneDrive(rcloneRemoteName, rcloneGDriveFolderID, "backup rclone"),
		maxBackupsToStore:     maxBackupsToStore,
		shutdownFunc:          shutdownFunc,
	}
}

// ListBackups lists available backups.
func (s *Backup) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(s.backupdir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("list backups: read dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()

		// case-insensitive check for .zip, but preserve original name
		if strings.EqualFold(filepath.Ext(n), ".zip") {
			names = append(names, strings.TrimSuffix(n, filepath.Ext(n)))
		}
	}

	sort.Slice(names, func(i, j int) bool {
		return names[i] > names[j] // timestamped names sort naturally
	})

	return names, nil
}

// SetCheckpointFunc sets the function to run before copying the DB into a backup (e.g. PRAGMA wal_checkpoint(FULL)).
// Call this after the DB is opened so backups include all committed data when SQLite is in WAL mode.
func (s *Backup) SetCheckpointFunc(f func() error) {
	s.checkpointDB = f
}

// CreateBackup creates a new backup of the current system state.
func (s *Backup) CreateBackup(syncToDrive bool) (string, error) {
	if err := s.ensureBackupDir(); err != nil {
		return "", err
	}

	ts := time.Now().UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("%s%s.zip", backupZipPrefix, ts)
	dst := filepath.Join(s.backupdir, name)

	log.Printf("creating backup in directory: %s", s.backupdir)
	// Overwrite protection: if exists, add a monotonic suffix.
	for fileExists(dst) {
		return "", fmt.Errorf("backup: file already exists: %s", dst)
	}

	f, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create backup: create zip: %w", err)
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	defer func() { _ = zw.Close() }()

	// Keep paths stable inside the ZIP, independent of absolute paths.
	// Layout:
	//   base_data/...    (contents of baseDataDir)
	//   models/...       (contents of modelsDir)
	//   items_metadata/... (contents of itemsMetadataDir)
	//   db/<basename>    (dbPath file)
	log.Printf("adding base data from %s to backup zip %s", s.baseDataDir, dst)
	if err := addDirToZip(zw, s.baseDataDir, "base_data"); err != nil {
		return "", err
	}
	log.Printf("adding models from %s to backup zip %s", s.modelsDir, dst)
	if err := addDirToZip(zw, s.modelsDir, "models"); err != nil {
		return "", err
	}
	log.Printf("adding items metadata from %s to backup zip %s", s.itemsMetadataDir, dst)
	if err := addDirToZip(zw, s.itemsMetadataDir, "items_metadata"); err != nil {
		return "", err
	}
	// Flush WAL into the main DB file before copying, so the backup contains all data.
	// Without this, on systems using WAL mode only the main file is copied and recent writes are in the -wal file.
	if s.checkpointDB != nil {
		log.Printf("running checkpoint function before adding DB to backup zip %s", dst)
		if err := s.checkpointDB(); err != nil {
			return "", fmt.Errorf("create backup: checkpoint db: %w", err)
		}
	}
	log.Printf("adding db from %s to backup zip %s", s.dbPath, dst)
	if err := addFileToZip(zw, s.dbPath, filepath.Join("db", filepath.Base(s.dbPath))); err != nil {
		return "", err
	}

	log.Printf("backup created: %s", dst)

	if err := s.ensureMaxBackups(); err != nil {
		log.Printf("warning: failed to ensure max backups after creating backup from zip: %v", err)
	}

	if syncToDrive {
		if err := s.syncBackupToDrive(dst, name); err != nil {
			return "", fmt.Errorf("create backup: sync to drive: %w", err)
		}
	}

	return strings.TrimSuffix(name, ".zip"), nil
}

func (s *Backup) syncBackupToDrive(localPath, filename string) error {
	if s.rclone.RemoteName == "" {
		return errors.New("backup: rclone remote name is not configured")
	}
	if _, err := s.rclone.Run("copy", localPath, s.rclone.RemotePath(filename)); err != nil {
		return err
	}
	if err := s.ensureMaxDriveBackups(); err != nil {
		return err
	}
	return nil
}

func (s *Backup) SyncBackupToDrive(backupId string) error {
	zipPath, err := s.GetBackupFullPath(backupId)
	if err != nil {
		return err
	}
	if err := s.syncBackupToDrive(zipPath, filepath.Base(zipPath)); err != nil {
		return fmt.Errorf("sync backup to drive: %w", err)
	}
	log.Printf("completed sync backup to drive: %s", zipPath)
	return nil
}

func (s *Backup) ensureMaxDriveBackups() error {
	entries, err := s.listDriveBackups()
	if err != nil {
		return fmt.Errorf("drive retention: list backups: %w", err)
	}
	if len(entries) <= s.maxBackupsToStore {
		return nil
	}

	log.Printf("deleting %d old backups from drive to enforce maxBackupsToStore=%d", len(entries)-s.maxBackupsToStore, s.maxBackupsToStore)
	for _, entry := range entries[s.maxBackupsToStore:] {
		if err := s.deleteDriveBackup(entry); err != nil {
			return fmt.Errorf("drive retention: delete %s: %w", entry.name, err)
		}
	}
	log.Printf("backups deleted from drive: %d", len(entries)-s.maxBackupsToStore)
	return nil
}

func (s *Backup) listDriveBackups() ([]driveBackupEntry, error) {
	out, err := s.rclone.Run("lsf", s.rclone.RemotePath(""))
	if err != nil {
		return nil, err
	}
	entries := parseDriveBackupEntries(string(out))
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name > entries[j].name
	})
	return entries, nil
}

func parseDriveBackupEntries(listing string) []driveBackupEntry {
	var entries []driveBackupEntry
	for _, line := range strings.Split(listing, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isDir := strings.HasSuffix(line, "/")
		name := strings.TrimSuffix(line, "/")
		if filepath.Base(name) != name {
			continue
		}
		if !strings.HasPrefix(name, backupZipPrefix) || !strings.EqualFold(filepath.Ext(name), ".zip") {
			continue
		}
		entries = append(entries, driveBackupEntry{name: name, isDir: isDir})
	}
	return entries
}

func (s *Backup) deleteDriveBackup(entry driveBackupEntry) error {
	if entry.isDir {
		_, err := s.rclone.Run("purge", s.rclone.RemotePath(entry.name))
		return err
	}
	_, err := s.rclone.Run("deletefile", s.rclone.RemotePath(entry.name))
	return err
}

// SetupRestoreBackup restores the system state from the backup with the given ID (or latest if backupId is empty or "latest").
func (s *Backup) SetupRestoreBackup(backupId string) error {
	backups, err := s.ListBackups()
	if err != nil {
		return err
	}
	if len(backups) == 0 {
		return errors.New("setup restore: no backups found")
	}

	backupToRestore := backupId
	if backupToRestore == "" || backupToRestore == "latest" {
		backupToRestore = backups[0]
	} else {
		if !lo.Contains(backups, backupToRestore) {
			return fmt.Errorf("setup restore: backup not found: %s", backupToRestore)
		}
	}

	log.Printf("backup: selected backup for restore: %s", backupToRestore)

	latest := filepath.Join(s.backupdir, backupToRestore+".zip")
	if !fileExists(latest) {
		return fmt.Errorf("setup restore: backup file not found: %s", latest)
	}

	// Extract into restoreFromBackupPath (clean first).
	if s.restoreFromBackupPath == "" {
		return errors.New("setup restore: restoreFromBackupPath is empty")
	}
	if err := os.RemoveAll(s.restoreFromBackupPath); err != nil {
		return fmt.Errorf("setup restore: clean restore dir: %w", err)
	}
	if err := os.MkdirAll(s.restoreFromBackupPath, 0o755); err != nil {
		return fmt.Errorf("setup restore: create restore dir: %w", err)
	}

	if err := unzip(latest, s.restoreFromBackupPath); err != nil {
		return fmt.Errorf("setup restore: unzip latest backup: %w", err)
	}

	log.Printf("backup: extracted latest backup %s to %s", latest, s.restoreFromBackupPath)

	// Shutdown so restore can happen safely on next start.
	if s.shutdownFunc != nil {
		log.Printf("backup: shutdown function")
		time.Sleep(5 * time.Second) // Give some time for logs to flush etc.
		if err := s.shutdownFunc(); err != nil {
			return fmt.Errorf("setup restore: shutdown: %w", err)
		}
	}
	return nil
}

func (s *Backup) RestoreLatestBackupIfRelevant() error {
	if s.restoreFromBackupPath == "" {
		return nil
	}
	if !dirExists(s.restoreFromBackupPath) {
		log.Printf("backup: restore from backup not found, skipping restore: %s", s.restoreFromBackupPath)
		return nil
	}

	// A valid extracted backup should contain expected top-level dirs.
	baseSrc := filepath.Join(s.restoreFromBackupPath, "base_data")
	modelsSrc := filepath.Join(s.restoreFromBackupPath, "models")
	itemsSrc := filepath.Join(s.restoreFromBackupPath, "items_metadata")
	dbSrcDir := filepath.Join(s.restoreFromBackupPath, "db")

	if !dirExists(baseSrc) || !dirExists(modelsSrc) || !dirExists(itemsSrc) || !dirExists(dbSrcDir) {
		return fmt.Errorf("restore: invalid restore directory structure at %s", s.restoreFromBackupPath)
	}

	// DB: take the first file found in db/ (should be exactly one).
	dbFile, err := firstRegularFile(dbSrcDir)
	if err != nil {
		return fmt.Errorf("restore: db file: %w", err)
	}

	log.Printf("restore: restoring base data from %s to %s", baseSrc, s.baseDataDir)
	if err := replaceDir(s.baseDataDir, baseSrc); err != nil {
		return fmt.Errorf("restore: base data: %w", err)
	}

	log.Printf("restore: restoring models from %s to %s", modelsSrc, s.modelsDir)
	if err := replaceDir(s.modelsDir, modelsSrc); err != nil {
		return fmt.Errorf("restore: models: %w", err)
	}

	log.Printf("restore: restoring items metadata from %s to %s", itemsSrc, s.itemsMetadataDir)
	if err := replaceDir(s.itemsMetadataDir, itemsSrc); err != nil {
		return fmt.Errorf("restore: items metadata: %w", err)
	}

	log.Printf("restore: restoring db from %s to %s", dbFile, s.dbPath)
	if err := replaceFile(s.dbPath, dbFile); err != nil {
		return fmt.Errorf("restore: db: %w", err)
	}

	// Cleanup restore payload.
	if err := os.RemoveAll(s.restoreFromBackupPath); err != nil {
		log.Printf("restore: warning: failed cleaning restore dir %s: %v", s.restoreFromBackupPath, err)
	} else {
		log.Printf("restore: cleaned restore dir %s", s.restoreFromBackupPath)
	}

	log.Printf("restore: completed")
	return nil
}

func (s *Backup) ensureBackupDir() error {
	if s.backupdir == "" {
		return errors.New("backup: backupdir is not set (empty string, check configuration)")
	}
	if err := os.MkdirAll(s.backupdir, 0o755); err != nil {
		return fmt.Errorf("backup: create backupdir: %w", err)
	}
	return nil
}

func (s *Backup) CreateBackupFromZip(file multipart.File, f func(dstPath string) error) (string, error) {
	dst, err := futils.CreateTemp("upload-*.zip")
	if err != nil {
		return "", fmt.Errorf("create backup from zip: create temp file: %w", err)
	}
	defer func() { _ = dst.Close() }()
	defer os.Remove(dst.Name())

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("create backup from zip: save file: %w", err)
	}

	if err := f(dst.Name()); err != nil {
		return "", fmt.Errorf("create backup from zip: process zip: %w", err)
	}

	// validations can be added here if needed, e.g. check for expected files/dirs in the extracted content.

	if err := futils.CopyFile(dst.Name(), filepath.Join(s.backupdir, fmt.Sprintf("%s%s.zip", backupZipPrefix, time.Now().UTC().Format("20060102T150405Z")))); err != nil {
		return "", fmt.Errorf("create backup from zip: save backup: %w", err)
	}

	if err := s.ensureMaxBackups(); err != nil {
		log.Printf("warning: failed to ensure max backups after creating backup from zip: %v", err)
	}

	return "backup created from zip", nil
}

func (s *Backup) GetBackupFullPath(backupId string) (string, error) {
	l, err := s.ListBackups()
	if err != nil {
		return "", err
	}
	if len(l) == 0 {
		return "", errors.New("no backups found")
	}
	if backupId == "latest" {
		backupId = l[0]
	}
	if !lo.Contains(l, backupId) {
		return "", fmt.Errorf("backup not found: %s", backupId)
	}
	return filepath.Join(s.backupdir, backupId+".zip"), nil
}

func (s *Backup) ensureMaxBackups() error {
	l, err := s.ListBackups()
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}
	if len(l) > s.maxBackupsToStore {
		toDelete := l[s.maxBackupsToStore:]
		for _, b := range toDelete {
			p := filepath.Join(s.backupdir, b+".zip")
			if err := os.Remove(p); err != nil {
				log.Printf("warning: failed to remove old backup %s: %v", p, err)
			} else {
				log.Printf("removed old backup %s", p)
			}
		}
	}
	return nil
}

// todo: the following should be in futils, some duplicate code...

func addDirToZip(zw *zip.Writer, srcDir, zipPrefix string) error {
	return addDirToZipWithSeen(zw, srcDir, zipPrefix, map[string]struct{}{})
}

func addDirToZipWithSeen(zw *zip.Writer, srcDir, zipPrefix string, seen map[string]struct{}) error {
	if srcDir == "" {
		return fmt.Errorf("zip: srcDir empty for %s", zipPrefix)
	}
	info, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Treat missing dirs as empty.
			return nil
		}
		return fmt.Errorf("zip: stat dir %s: %w", srcDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("zip: expected dir, got file: %s", srcDir)
	}
	canonicalDir, err := filepath.EvalSymlinks(srcDir)
	if err != nil {
		return fmt.Errorf("zip: eval symlinks %s: %w", srcDir, err)
	}
	if _, ok := seen[canonicalDir]; ok {
		log.Printf("zip: skipping already visited directory %s as %s", srcDir, zipPrefix)
		return nil
	}
	seen[canonicalDir] = struct{}{}

	statsSeen := copySeenDirs(seen)
	stats, err := collectDirZipStats(srcDir, statsSeen)
	if err != nil {
		return err
	}
	log.Printf("zip: adding directory %s as %s: %d files, %d directories, %d bytes (%s)", srcDir, zipPrefix, stats.files, stats.dirs, stats.size, futils.FormatBytes(stats.size))
	if stats.files == 0 {
		log.Printf("zip: directory %s as %s has no files to add", srcDir, zipPrefix)
	}

	addedFiles := 0
	addedBytes := int64(0)
	err = filepath.WalkDir(canonicalDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("zip: walk %s: %w", srcDir, walkErr)
		}

		rel, err := filepath.Rel(canonicalDir, path)
		if err != nil {
			return fmt.Errorf("zip: rel path: %w", err)
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			// Explicit directory entries are optional, but harmless.
			if rel == "." {
				return nil
			}
			h := &zip.FileHeader{
				Name:   filepath.ToSlash(filepath.Join(zipPrefix, rel)) + "/",
				Method: zip.Deflate,
			}
			h.SetMode(0o755)
			_, err := zw.CreateHeader(h)
			return err
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("zip: stat file %s: %w", path, err)
		}
		if fileInfo.IsDir() {
			if d.Type()&os.ModeSymlink == 0 {
				return fmt.Errorf("zip: expected file, got dir: %s", path)
			}
			return addDirToZipWithSeen(zw, path, filepath.ToSlash(filepath.Join(zipPrefix, rel)), seen)
		}

		if err := addFileToZip(zw, path, filepath.ToSlash(filepath.Join(zipPrefix, rel))); err != nil {
			return err
		}
		addedFiles++
		addedBytes += fileInfo.Size()
		if addedFiles%zipProgressLogEveryFiles == 0 {
			log.Printf("zip: added %d/%d files from %s (%d/%d bytes, %s/%s)", addedFiles, stats.files, srcDir, addedBytes, stats.size, futils.FormatBytes(addedBytes), futils.FormatBytes(stats.size))
		}
		return nil
	})
	if err != nil {
		return err
	}

	log.Printf("zip: finished adding %s as %s: %d/%d files, %d/%d bytes (%s/%s)", srcDir, zipPrefix, addedFiles, stats.files, addedBytes, stats.size, futils.FormatBytes(addedBytes), futils.FormatBytes(stats.size))
	return nil
}

func collectDirZipStats(srcDir string, seen map[string]struct{}) (dirZipStats, error) {
	var stats dirZipStats
	canonicalDir, err := filepath.EvalSymlinks(srcDir)
	if err != nil {
		return dirZipStats{}, fmt.Errorf("zip: eval symlinks %s: %w", srcDir, err)
	}
	err = filepath.WalkDir(canonicalDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("zip: scan %s: %w", srcDir, walkErr)
		}
		if d.IsDir() {
			if path != canonicalDir {
				stats.dirs++
			}
			return nil
		}

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("zip: stat file %s: %w", path, err)
		}
		if info.IsDir() {
			if d.Type()&os.ModeSymlink == 0 {
				return fmt.Errorf("zip: expected file, got dir: %s", path)
			}
			canonicalDir, err := filepath.EvalSymlinks(path)
			if err != nil {
				return fmt.Errorf("zip: eval symlinks %s: %w", path, err)
			}
			if _, ok := seen[canonicalDir]; ok {
				return nil
			}
			seen[canonicalDir] = struct{}{}
			nestedStats, err := collectDirZipStats(canonicalDir, seen)
			if err != nil {
				return err
			}
			stats.files += nestedStats.files
			stats.dirs += nestedStats.dirs + 1
			stats.size += nestedStats.size
			return nil
		}
		stats.files++
		stats.size += info.Size()
		return nil
	})
	if err != nil {
		return dirZipStats{}, err
	}
	return stats, nil
}

func copySeenDirs(seen map[string]struct{}) map[string]struct{} {
	cp := make(map[string]struct{}, len(seen))
	for k, v := range seen {
		cp[k] = v
	}
	return cp
}

func addFileToZip(zw *zip.Writer, srcPath, zipPath string) error {
	if srcPath == "" {
		return errors.New("zip: srcPath empty")
	}
	info, err := os.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Treat missing file as absent.
			return nil
		}
		return fmt.Errorf("zip: stat file %s: %w", srcPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("zip: expected file, got dir: %s", srcPath)
	}

	h, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("zip: header %s: %w", srcPath, err)
	}
	h.Name = filepath.ToSlash(zipPath)
	h.Method = zip.Deflate

	w, err := zw.CreateHeader(h)
	if err != nil {
		return fmt.Errorf("zip: create %s: %w", zipPath, err)
	}

	f, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("zip: open %s: %w", srcPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("zip: copy %s: %w", srcPath, err)
	}
	return nil
}

func unzip(zipPath, dstDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("unzip: open %s: %w", zipPath, err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		name := f.Name
		if strings.Contains(name, "..") {
			return fmt.Errorf("unzip: suspicious path %q", name)
		}

		outPath := filepath.Join(dstDir, filepath.FromSlash(name))
		if !strings.HasPrefix(outPath, filepath.Clean(dstDir)+string(os.PathSeparator)) {
			return fmt.Errorf("unzip: path traversal blocked: %q", name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0o755); err != nil {
				return fmt.Errorf("unzip: mkdir %s: %w", outPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("unzip: mkdir parent %s: %w", outPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("unzip: open entry %s: %w", name, err)
		}

		tmp := outPath + ".tmp"
		w, err := os.Create(tmp)
		if err != nil {
			_ = rc.Close()
			return fmt.Errorf("unzip: create %s: %w", tmp, err)
		}

		_, copyErr := io.Copy(w, rc)
		closeErr := w.Close()
		_ = rc.Close()

		if copyErr != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("unzip: write %s: %w", outPath, copyErr)
		}
		if closeErr != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("unzip: close %s: %w", outPath, closeErr)
		}

		if err := os.Rename(tmp, outPath); err != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("unzip: rename %s: %w", outPath, err)
		}
	}

	return nil
}

func replaceDir(dstDir, srcDir string) error {
	// Remove dst then copy src -> dst.
	if err := os.RemoveAll(dstDir); err != nil {
		return fmt.Errorf("replace dir: remove %s: %w", dstDir, err)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("replace dir: mkdir %s: %w", dstDir, err)
	}
	return copyDir(srcDir, dstDir)
}

func copyDir(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("copy dir: walk %s: %w", srcDir, walkErr)
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("copy dir: rel: %w", err)
		}
		if rel == "." {
			return nil
		}

		dstPath := filepath.Join(dstDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		return copyFile(path, dstPath)
	})
}

func replaceFile(dstFile, srcFile string) error {
	if err := os.MkdirAll(filepath.Dir(dstFile), 0o755); err != nil {
		return fmt.Errorf("replace file: mkdir parent: %w", err)
	}
	tmp := dstFile + ".tmp"
	if err := copyFile(srcFile, tmp); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dstFile); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace file: rename: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copy file: open src %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("copy file: stat src %s: %w", src, err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("copy file: create dst %s: %w", dst, err)
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()

	if copyErr != nil {
		return fmt.Errorf("copy file: copy to %s: %w", dst, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("copy file: close dst %s: %w", dst, closeErr)
	}

	// Best-effort permissions.
	_ = os.Chmod(dst, info.Mode())
	return nil
}

func firstRegularFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		p := filepath.Join(dir, e.Name())
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		if fi.Mode().IsRegular() {
			return p, nil
		}
	}
	return "", fmt.Errorf("no regular file found in %s", dir)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func dirExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}
